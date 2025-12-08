package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"sync/atomic"
)

const DEFAULT_PORT = 8080
const STATIC_DIR = "./static"
const STATIC_ASSETS_DIR = STATIC_DIR + "/assets"

type apiConfig struct {
	FileserverHits atomic.Int32
	Port int
}

func getServerPort() (int, error) {
	portEnv := os.Getenv("CHIRPY_PORT")

	if portEnv == "" {
		return -1, fmt.Errorf("no port defined")
	}

	port, err := strconv.Atoi(portEnv)
	if err != nil {
		return -1, fmt.Errorf("invalid port \"%s\": %w", portEnv, err)
	}

	return port, nil
}

func middlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: %v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

func (cfg *apiConfig) middlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)
	})
}

var handleHealthCheck = middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}))


func (cfg *apiConfig) middlewareHandleMetrics() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := fmt.Sprintf("Hits: %d", cfg.FileserverHits.Load())
	
		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}))
}

func (cfg *apiConfig) middlewareHandleReset() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cfg.FileserverHits.Store(0)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric has been reset"))
	}))
}

var handleHome = middlewareLogger(http.StripPrefix("/app", http.FileServer(http.Dir(STATIC_DIR))))
var handleAssets = middlewareLogger(http.StripPrefix("/app/assets", http.FileServer(http.Dir(STATIC_ASSETS_DIR))))

func main() {
	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, DEFAULT_PORT)
		port = DEFAULT_PORT
	}

	cfg := apiConfig{
		FileserverHits: atomic.Int32{},
		Port: port,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /metrics", cfg.middlewareHandleMetrics())
	mux.Handle("POST /reset", cfg.middlewareHandleReset())
	mux.Handle("GET /healthz", handleHealthCheck)
	mux.Handle("/app/assets/", handleAssets)
	mux.Handle("/app/", cfg.middlewareMetricsInc(handleHome))

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := http.Server{
		Addr: addr,
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("could not start the server: %v", err)
	}

	fmt.Printf("server start on port: %d\n", port)
}
