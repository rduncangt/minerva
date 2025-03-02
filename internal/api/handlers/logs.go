package handlers

import (
	"database/sql"
	"net/http"
	"strconv"

	"minerva/internal/api"
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
			api.JsonErrorResponse(w, http.StatusInternalServerError, "Database error")
			return
		}
		defer rows.Close()

		logs := []map[string]interface{}{}
		for rows.Next() {
			var ts, srcIP, dstIP, proto, action string
			if err := rows.Scan(&ts, &srcIP, &dstIP, &proto, &action); err != nil {
				api.JsonErrorResponse(w, http.StatusInternalServerError, "Scan error")
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

		api.JsonResponse(w, http.StatusOK, map[string]interface{}{"data": logs})
	}
}
