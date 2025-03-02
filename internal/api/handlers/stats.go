package handlers

import (
	"database/sql"
	"net/http"

	"minerva/internal/api"
)

// GetStats returns database size and row counts.
func GetStats(db *sql.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var dbSize string
		if err := db.QueryRow("SELECT pg_size_pretty(pg_database_size(current_database()))").Scan(&dbSize); err != nil {
			api.JsonErrorResponse(w, http.StatusInternalServerError, "Failed to get database size")
			return
		}

		rows, err := db.Query(`
			SELECT c.relname AS table_name, 
				COALESCE(s.n_live_tup, 0) AS row_count, 
				pg_size_pretty(pg_total_relation_size(c.oid)) AS size 
			FROM pg_class c
			JOIN pg_stat_user_tables s ON c.relname = s.relname
			ORDER BY pg_total_relation_size(c.oid) DESC;
		`)
		if err != nil {
			api.JsonErrorResponse(w, http.StatusInternalServerError, "Failed to get table statistics")
			return
		}
		defer rows.Close()

		tables := []map[string]interface{}{}
		for rows.Next() {
			var tableName string
			var rowCount int64
			var size string
			if err := rows.Scan(&tableName, &rowCount, &size); err != nil {
				api.JsonErrorResponse(w, http.StatusInternalServerError, "Failed to scan table statistics")
				return
			}
			tables = append(tables, map[string]interface{}{
				"name":      tableName,
				"row_count": rowCount,
				"size":      size,
			})
		}

		api.JsonResponse(w, http.StatusOK, map[string]interface{}{
			"data": map[string]interface{}{
				"database_size": dbSize,
				"tables":        tables,
			},
		})
	}
}
