package api

import (
	"database/sql"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
)

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
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		defer rows.Close()

		logs := []map[string]interface{}{}
		for rows.Next() {
			var ts, srcIP, dstIP, proto, action string
			if err := rows.Scan(&ts, &srcIP, &dstIP, &proto, &action); err != nil {
				http.Error(w, "Scan error", http.StatusInternalServerError)
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

		json.NewEncoder(w).Encode(logs)
	}
}

// GetStats returns basic statistics from the log_data table.
func GetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		query := `SELECT COUNT(*) AS total_logs FROM log_data`
		var totalLogs int
		if err := db.QueryRow(query).Scan(&totalLogs); err != nil {
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
		stats := map[string]interface{}{"total_logs": totalLogs}
		json.NewEncoder(w).Encode(stats)
	}
}

// GetGeo returns geolocation data for an IP address.
func GetGeo(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		ip := vars["ip"]
		query := `SELECT country, region, city, isp FROM ip_geo WHERE ip_address = $1`
		var country, region, city, isp string
		if err := db.QueryRow(query, ip).Scan(&country, &region, &city, &isp); err != nil {
			http.Error(w, "IP not found", http.StatusNotFound)
			return
		}
		geoData := map[string]string{"ip": ip, "country": country, "region": region, "city": city, "isp": isp}
		json.NewEncoder(w).Encode(geoData)
	}
}
