package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
)

const DEFAULT_PORT = 8080
const STATIC_DIR = "./static"
const STATIC_ASSETS_DIR = STATIC_DIR + "/assets"
var PROFANE_WORDS []string = []string{"kerfuffle", "sharbert", "fornax"}

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

func cleanChirp(textToClean string, profaneWords []string) string {
	chunks := strings.Split(textToClean, " ")

	for cIndex, c := range chunks {
		for _, word := range profaneWords {
			if strings.EqualFold(c, word) {
				chunks[cIndex] = strings.Repeat("*", 4)
			}
		}
	}

	return strings.Join(chunks, " ")
}

func respondWithPlainText(w http.ResponseWriter, statusCode int, text string) {
	w.Header().Add("Content-Type", "text/plain; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write([]byte(text))
}

func respondWithJSON(w http.ResponseWriter, statusCode int, data []byte) {
	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(data)
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

var handleValidateChirp = middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	type validateChirpResource struct {
		Body string `json:"body"`
	}
	type validateChirpResponse struct {
		Error string `json:"error,omitempty"`
		Valid bool `json:"valid"`
	}

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	var validatedChirp validateChirpResource
	err := decoder.Decode(&validatedChirp)

	if err != nil {		
		payload := validateChirpResponse{
			Error: "Something went wrong",
			Valid: false,
		}
		
		data, err := json.Marshal(payload)
		if err != nil {
			errorMsg := fmt.Sprintf("error marshalling response: %v", err)

			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorMsg))
			return
		}
		
		w.Header().Add("Content-Type", "text/json; charset=utf-8")
		w.WriteHeader(http.StatusBadRequest)
		w.Write(data)
		return
	}

	if len(validatedChirp.Body) > 140 {
		payload := validateChirpResponse{
			Error: "Chirp is too long",
			Valid: false,
		}
		
		data, err := json.Marshal(payload)
		if err != nil {
			errorMsg := fmt.Sprintf("error marshalling response: %v", err)
			
			w.Header().Add("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(errorMsg))
			return
		}
		
		w.WriteHeader(http.StatusBadRequest)
		w.Header().Add("Content-Type", "text/json; charset=utf-8")
		w.Write(data)
		return
	}

	payload := validateChirpResponse{
		Error: "",
		Valid: true,
	}

	data, err := json.Marshal(payload)
	if err != nil {
		errorMsg := fmt.Sprintf("error marshalling response: %v", err)

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(errorMsg))
		return
	}

	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	w.Write(data)
}))

func (cfg *apiConfig) middlewareHandleMetrics() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := fmt.Sprintf(`
			<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>`,
			cfg.FileserverHits.Load())
	
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
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

	mux.Handle("GET /admin/metrics", cfg.middlewareHandleMetrics())
	mux.Handle("POST /admin/reset", cfg.middlewareHandleReset())
	mux.Handle("GET /api/healthz", handleHealthCheck)
	mux.Handle("POST /api/validate_chirp", handleValidateChirp)
	mux.Handle("/app/assets/", cfg.middlewareMetricsInc(handleAssets))
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
