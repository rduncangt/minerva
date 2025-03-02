package handlers

import (
	"database/sql"
	"net/http"

	"minerva/internal/api"

	"github.com/gorilla/mux"
)

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
			api.JsonErrorResponse(w, http.StatusNotFound, "IP not found")
			return
		}

		geoData := map[string]interface{}{
			"ip":        ip,
			"country":   country,
			"region":    region,
			"city":      city,
			"isp":       isp,
			"latitude":  latitude.Float64,
			"longitude": longitude.Float64,
		}

		api.JsonResponse(w, http.StatusOK, map[string]interface{}{"data": geoData})
	}
}
