package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
	"github.com/nacen-dev/chirpy/internal/database"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	jwtSecret      string
}

func main() {
	godotenv.Load()
	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	dbConnection, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatal("Error connecting to the database")
	}
	dbQueries := database.New(dbConnection)

	const filepathRoot = "."
	const port = "8080"

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}
	apiCfg := apiConfig{
		fileserverHits: atomic.Int32{},
		db:             dbQueries,
		platform:       platform,
		jwtSecret:      jwtSecret,
	}

	serveMux := http.NewServeMux()
	serveMux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(filepathRoot)))))

	serveMux.HandleFunc("POST /admin/reset", apiCfg.handleResetUsers)
	serveMux.HandleFunc("GET /admin/metrics", apiCfg.handleNumberOfRequest)

	serveMux.HandleFunc("GET /api/healthz", handleHealthCheck)
	serveMux.HandleFunc("POST /api/login", apiCfg.handleLogin)
	serveMux.HandleFunc("POST /api/refresh", apiCfg.handleRefresh)
	serveMux.HandleFunc("POST /api/revoke", apiCfg.handleRevoke)
	serveMux.HandleFunc("POST /api/polka/webhooks", apiCfg.handleUpgradeToChirpyRed)

	serveMux.HandleFunc("GET /api/chirps", apiCfg.handleGetChirps)
	serveMux.HandleFunc("GET /api/chirps/{chirpId}", apiCfg.handleGetChirpById)
	serveMux.HandleFunc("DELETE /api/chirps/{chirpId}", apiCfg.handleDeleteChirpById)
	serveMux.HandleFunc("POST /api/chirps", apiCfg.handleCreateChirp)

	serveMux.HandleFunc("POST /api/users", apiCfg.handleCreateUsers)
	serveMux.HandleFunc("PUT /api/users", apiCfg.handleUpdateUser)

	server := http.Server{
		Addr:    ":" + port,
		Handler: serveMux,
	}

	log.Printf("Serving files from %s on port: %s\n", filepathRoot, port)
	log.Fatal(server.ListenAndServe())
}
