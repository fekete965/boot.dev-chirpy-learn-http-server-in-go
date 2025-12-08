package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
)

const DEFAULT_PORT = 8080
const STATIC_DIR = "./static"
const STATIC_ASSETS_DIR = STATIC_DIR + "/assets"

type ServerConfig struct {
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

var handleHealthCheck = middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}))

var handleHome = middlewareLogger(http.StripPrefix("/app", http.FileServer(http.Dir(STATIC_DIR))))
var handleAssets = middlewareLogger(http.StripPrefix("/app/assets/", http.FileServer(http.Dir(STATIC_ASSETS_DIR))))

func main() {
	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, DEFAULT_PORT)
		port = DEFAULT_PORT
	}

	cfg := ServerConfig{
		Port: port,
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/healthz/", handleHealthCheck)
	mux.Handle("/app/assets/", handleAssets)
	mux.Handle("/app", handleHome)

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
