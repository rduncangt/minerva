package main

import (
	"log"
	"minerva/internal/api"
	"minerva/internal/config"
	"minerva/internal/db"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	log.SetOutput(os.Stderr)
	log.Println("Starting minerva-api...")

	// Load configuration
	conf, err := config.LoadConfig("minerva_config.toml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Override database name if environment variable is set
	if dbName := os.Getenv("MINERVA_DB_NAME"); dbName != "" {
		conf.Database.Name = dbName
	}

	// Connect to the database
	dbPort := os.Getenv("MINERVA_DB_PORT")
	if dbPort == "" {
		dbPort = "5432" // Default PostgreSQL port
	}
	database, err := db.Connect(conf.Database.Host, dbPort, conf.Database.User, conf.Database.Password, conf.Database.Name)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	log.Println("minerva-api is running and connected to the database.")

	// Set up HTTP server (placeholder)
	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	router := mux.NewRouter()
	router.HandleFunc("/api/v1/logs", api.GetLogs(database)).Methods("GET")
	router.HandleFunc("/api/v1/stats", api.GetStats(database)).Methods("GET")
	router.HandleFunc("/api/v1/geo/{ip}", api.GetGeo(database)).Methods("GET")

	log.Fatal(http.ListenAndServe(":8080", nil))
}
