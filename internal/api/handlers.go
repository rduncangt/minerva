package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

// jsonResponse writes JSON response with appropriate headers.
func jsonResponse(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

// jsonErrorResponse writes a JSON error response.
func jsonErrorResponse(w http.ResponseWriter, status int, message string) {
	jsonResponse(w, status, map[string]string{"error": message})
}

// GetLogs returns a paginated list of logs from the log_data table.
func GetLogs(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit, err := strconv.Atoi(r.URL.Query().Get("limit"))
		if err != nil || limit <= 0 {
			limit = 50
		}
		offset, err := strconv.Atoi(r.URL.Query().Get("offset"))
		if err != nil || offset < 0 {
			offset = 0
		}

		query := `SELECT timestamp, source_ip, destination_ip, protocol, action FROM log_data ORDER BY timestamp DESC LIMIT $1 OFFSET $2`
		rows, err := db.Query(query, limit, offset)
		if err != nil {
			jsonErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		logs := []map[string]interface{}{}
		for rows.Next() {
			var ts, srcIP, dstIP, proto, action string
			if err := rows.Scan(&ts, &srcIP, &dstIP, &proto, &action); err != nil {
				jsonErrorResponse(w, http.StatusInternalServerError, "Scan error")
				return
			}
			logs = append(logs, map[string]interface{}{
				"timestamp":      ts,
				"source_ip":      srcIP,
				"destination_ip": dstIP,
				"protocol":       proto,
				"action":         action,
			})
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{"data": logs})
	}
}

// GetStats returns basic statistics from the log_data table.
func GetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT COUNT(*) AS total_logs FROM log_data`
		var totalLogs int
		if err := db.QueryRow(query).Scan(&totalLogs); err != nil {
			jsonErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}
		jsonResponse(w, http.StatusOK, map[string]interface{}{"data": map[string]interface{}{"total_logs": totalLogs}})
	}
}

// GetGeo returns geolocation data for an IP address.
func GetGeo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ip := vars["ip"]
		query := `SELECT country, region, city, isp, latitude, longitude FROM ip_geo WHERE ip_address = $1`
		var country, region, city, isp string
		var latitude, longitude sql.NullFloat64

		err := db.QueryRow(query, ip).Scan(&country, &region, &city, &isp, &latitude, &longitude)
		if err != nil {
			jsonErrorResponse(w, http.StatusNotFound, "IP not found")
			return
		}

		geoData := map[string]interface{}{
			"ip":        ip,
			"country":   country,
			"region":    region,
			"city":      city,
			"isp":       isp,
			"latitude":  latitude,
			"longitude": longitude,
		}

		jsonResponse(w, http.StatusOK, map[string]interface{}{"data": geoData})
	}
}
