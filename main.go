package main

import (
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/google/uuid"
	"github.com/joho/godotenv"
	"github.com/lib/pq"
)

const DEFAULT_PORT = 8080
const STATIC_DIR = "./static"
const STATIC_ASSETS_DIR = STATIC_DIR + "/assets"
var PROFANE_WORDS []string = []string{"kerfuffle", "sharbert", "fornax"}

type apiConfig struct {
	DbQueries *database.Queries
	FileserverHits atomic.Int32
	JWTSecret string
	Platform string
	Port int
}

func getServerPort() (int, error) {
	portEnv := os.Getenv("PORT")

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

func respondWithJSON(w http.ResponseWriter, statusCode int, data interface{}) {
	marshalledData, err := json.Marshal(data)
	if err != nil {
		errorMessage := fmt.Sprintf("error marshalling response: %v", err)
		
		respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
		return
	}

	w.Header().Add("Content-Type", "text/json; charset=utf-8")
	w.WriteHeader(statusCode)
	w.Write(marshalledData)
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
	respondWithPlainText(w, http.StatusOK, "OK")
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

		if cfg.Platform != "dev" {
			respondWithPlainText(w, http.StatusForbidden, "Forbidden operation")
			return
		}

		// Reset the metrics
		cfg.FileserverHits.Store(0)

		// Reset the database
		err := cfg.DbQueries.DeleteAllUsers(r.Context())
		if err != nil {
			errorMessage := fmt.Sprintf("error deleting all users: %v", err)
			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric has been reset"))
	}))
}

func (cfg *apiConfig) handleCreateUser() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type createUserResource struct {
			Email string `json:"email"`
			Password string `json:"password"`
		}
		
		type createUserResponse struct {
			Id uuid.UUID `json:"id"`
			Email string `json:"email"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var payload createUserResource
		err := decoder.Decode(&payload)
		if err != nil {
			respondWithPlainText(w, http.StatusBadRequest, "Invalid request body")
			return 
		}

		hashedPassword, err := auth.HashPassword(payload.Password)
		if err != nil {
			fmt.Printf("error hashing password: %v", err)
			errorMessage := "error during password handling"

			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		newUser, err := cfg.DbQueries.CreateUser(r.Context(), database.CreateUserParams{
			ID: uuid.New(),
			Email: payload.Email,
			HashedPassword: hashedPassword,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		if err != nil {
			var pqErr *pq.Error

			if errors.As(err, &pqErr) && pqErr.Code == "23505" {
				errorMessage := fmt.Sprintf("Email already in use: %s", payload.Email)
				respondWithPlainText(w, http.StatusConflict, errorMessage)
				return
			}

			errorMessage := fmt.Sprintf("Error creating user: %v", err)
			respondWithPlainText(w, http.StatusConflict, errorMessage)
			return
		}

		data := createUserResponse{
			Id: newUser.ID,
			Email: newUser.Email,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
		}

		respondWithJSON(w, http.StatusCreated, data)
	}))
}

func (cfg *apiConfig) handleCreateChirp() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type validateChirpResource struct {
			Body string `json:"body"`
			UserID string `json:"user_id"`
		}
	
		type validateChirpResponse struct {
			ID uuid.UUID `json:"id"`
			UserID uuid.UUID `json:"user_id"`
			Body string `json:"body"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		bearerToken, err := auth.GetBearerToken(r)
		if err != nil {
			errorMessage := fmt.Sprintf("error during authentication: %v", err)

			respondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}

		userID, err := auth.ValidateJWT(bearerToken, cfg.JWTSecret)
		if err != nil {
			errorMessage := fmt.Sprintf("error during authentication: %v", err)

			respondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}
	
		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()
	
		var validatedChirp validateChirpResource
		err := decoder.Decode(&validatedChirp)
	
		if err != nil {
			errorMessage := fmt.Sprintf("error decoding request body: %v", err)

			respondWithPlainText(w, http.StatusBadRequest, errorMessage)
			return
		}
	
		if len(validatedChirp.Body) > 140 {
			errorMessage := "Chirp is too long"

			respondWithPlainText(w, http.StatusBadRequest, errorMessage)
			return
		}
	
		cleanedBody := cleanChirp(validatedChirp.Body, PROFANE_WORDS)

		newChirp, err := cfg.DbQueries.CreateChirp(r.Context(), database.CreateChirpParams{
			ID: uuid.New(),
			UserID: uuid.MustParse(validatedChirp.UserID),
			Body: cleanedBody,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		})

		if err != nil {
			errorMessage := fmt.Sprintf("error creating chirp: %v", err)

			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		data := validateChirpResponse{
			ID: newChirp.ID,
			UserID: newChirp.UserID,
			Body: newChirp.Body,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
		}
		
		respondWithJSON(w, http.StatusCreated, data)
	}))
}

func (cfg *apiConfig) handleGetAllChirps() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type getAllChirpResponse struct {
			ID uuid.UUID `json:"id"`
			UserID uuid.UUID `json:"user_id"`
			Body string `json:"body"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		
		chirps, err := cfg.DbQueries.GetAllChirps(r.Context())
		if err != nil {
			errorMessage := fmt.Sprintf("error getting all chirps: %v", err)

			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		data := make([]getAllChirpResponse, len(chirps))
		for i, chirp := range chirps {
			data[i] = getAllChirpResponse{
				ID: chirp.ID,
				UserID: chirp.UserID,
				Body: chirp.Body,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
			}
		}

		respondWithJSON(w, http.StatusOK, data)
	}))
}

func safeParseUUID(str string) (uuid.UUID, error) {
	err := uuid.Validate(str)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID: %v", err)
	}

	return uuid.MustParse(str), nil
}

func (cfg *apiConfig) handleGetChirpById() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type getChirpByIdResponse struct {
			ID uuid.UUID `json:"id"`
			UserId uuid.UUID `json:"user_id"`
			Body string `json:"body"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}
		
		chirpID, err := safeParseUUID(r.PathValue("chirpID"))
		if err != nil {			
			errorMessage := fmt.Sprintf("cannot parse chirpID: %v", err)

			respondWithPlainText(w, http.StatusBadRequest, errorMessage)
			return
		}

		chirp, err := cfg.DbQueries.GetChirpById(r.Context(), chirpID)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errorMessage := fmt.Sprintf("chirp not found by id #%v", chirpID)

				respondWithPlainText(w, http.StatusNotFound, errorMessage)
				return
			}

			errorMessage := fmt.Sprintf("error getting chirp by id #%v: %v", chirpID, err)

			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		data := getChirpByIdResponse{
			ID: chirp.ID,
			UserId: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}

		respondWithJSON(w, http.StatusOK, data)
	}))
}

func (cfg *apiConfig) handleLogin() http.Handler {
	return middlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		type loginResource struct {
			Email string `json:"email"`
			Password string `json:"password"`
		}

		type loginResponse struct {
			ID uuid.UUID `json:"id"`
			Email string `json:"email"`
			CreatedAt time.Time `json:"created_at"`
			UpdatedAt time.Time `json:"updated_at"`
		}

		decoder := json.NewDecoder(r.Body)
		defer r.Body.Close()

		var payload loginResource
		err := decoder.Decode(&payload)
		if err != nil {
			errorMessage := fmt.Sprintf("error decoding request body: %v", err)

			respondWithPlainText(w, http.StatusBadGateway, errorMessage)
			return
		}

		user, err := cfg.DbQueries.FindUserByEmail(r.Context(), payload.Email)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				errorMessage := fmt.Sprintf("user not found by email: %s", payload.Email)
				
				respondWithPlainText(w, http.StatusNotFound, errorMessage)
				return
			}
			errorMessage := fmt.Sprintf("error getting user by email %v: %s", payload.Email, err)
				
			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		match, err := auth.CheckPasswordHash(payload.Password, user.HashedPassword)
		if err != nil {
			errorMessage := fmt.Sprintf("error during user validation: %v", err)

			respondWithPlainText(w, http.StatusInternalServerError, errorMessage)
			return
		}

		if !match {
			errorMessage := "invalid user credentials"

			respondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}

		data := loginResponse{
			ID: user.ID,
			Email: user.Email,
			CreatedAt: user.CreatedAt,
			UpdatedAt: user.UpdatedAt,
		}

		respondWithJSON(w, http.StatusOK, data)
	}))
}

var handleHome = middlewareLogger(http.StripPrefix("/app", http.FileServer(http.Dir(STATIC_DIR))))
var handleAssets = middlewareLogger(http.StripPrefix("/app/assets", http.FileServer(http.Dir(STATIC_ASSETS_DIR))))

type envVars struct {
	DbUrl string
	Platform string
	Port int
}

func loadEnv() (envVars, error) {
	err := godotenv.Load(".env")
	if err != nil {
		errorMessage := fmt.Errorf("error loading env file: %v", err)
		return envVars{}, errorMessage
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		errorMessage := fmt.Errorf("DB_URL is not set")
		return envVars{}, errorMessage
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		errorMessage := fmt.Errorf("PLATFORM is not set")
		return envVars{}, errorMessage
	}

	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, DEFAULT_PORT)
		port = DEFAULT_PORT
	}

	return envVars{
		DbUrl: dbUrl,
		Platform: platform,
		Port: port,
	}, nil
}

func main() {
	envVars, err := loadEnv()
	if err != nil {
		log.Fatalf("error loading environment variables: %v", err)
	}

	dbConnection, err := sql.Open("postgres", envVars.DbUrl)
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	cfg := apiConfig{
		DbQueries: database.New(dbConnection),
		FileserverHits: atomic.Int32{},
		Platform: envVars.Platform,
		Port: envVars.Port,
	}

	mux := http.NewServeMux()

	mux.Handle("GET /admin/metrics", cfg.middlewareHandleMetrics())
	mux.Handle("POST /admin/reset", cfg.middlewareHandleReset())
	mux.Handle("GET /api/healthz", handleHealthCheck)
	mux.Handle("POST /api/chirps", cfg.handleCreateChirp())
	mux.Handle("GET /api/chirps", cfg.handleGetAllChirps())
	mux.Handle("GET /api/chirps/{chirpID}", cfg.handleGetChirpById())
	mux.Handle("POST /api/users", cfg.handleCreateUser())
	mux.Handle("POST /api/login", cfg.handleLogin())
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

	fmt.Printf("server started on port: %d\n", envVars.Port)
}
