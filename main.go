package main

import (
	"database/sql"
	"log"
	"net/http"
	"os"
	"sync/atomic"

	"github.com/Kamciooo1226/chirpy/internal/database"
	"github.com/joho/godotenv"
	_ "github.com/lib/pq"
)

type apiConfig struct {
	fileserverHits atomic.Int32
	db             *database.Queries
	platform       string
	secret         string
}

func main() {
	godotenv.Load()

	dbURL := os.Getenv("DB_URL")
	if dbURL == "" {
		log.Fatal("DB_URL must be set")
	}
	platform := os.Getenv("PLATFORM")
	if platform == "" {
		log.Fatal("PLATFORM must be set")
	}
	secret := os.Getenv("SECRET")
	if secret == "" {
		log.Fatal("SECRET must be set")
	}
	db, err := sql.Open("postgres", dbURL)
	if err != nil {
		log.Fatalf("Error opening connection to the database: %s", err)
	}

	dbQueries := database.New(db)

	mux := http.NewServeMux()
	apiCfg := apiConfig{
		db:       dbQueries,
		platform: platform,
	}

	mux.Handle("/app/", apiCfg.middlewareMetricsInc(http.StripPrefix("/app", http.FileServer(http.Dir(".")))))

	mux.HandleFunc("GET /admin/metrics", apiCfg.handlerMetrics)
	mux.HandleFunc("POST /admin/reset", apiCfg.handlerDeleteUsers)

	mux.HandleFunc("GET /api/healthz", handlerReadiness)

	mux.HandleFunc("POST /api/chirps", apiCfg.handlerChirpsCreate)
	mux.HandleFunc("GET /api/chirps", apiCfg.handlerGetAllChirps)
	mux.HandleFunc("GET /api/chirps/{chirpID}", apiCfg.handlerGetChirp)

	mux.HandleFunc("POST /api/users", apiCfg.handlerCreateUser)
	mux.HandleFunc("POST /api/login", apiCfg.handlerUserLogin)

	server := &http.Server{
		Addr:    ":8080",
		Handler: mux,
	}
	err = server.ListenAndServe()
	if err != nil {
		log.Fatal(err)
	}
}
