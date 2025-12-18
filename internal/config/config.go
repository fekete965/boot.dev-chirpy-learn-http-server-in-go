package config

import (
	"fmt"
	"os"
	"strconv"
	"sync/atomic"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/joho/godotenv"
)

type ApiConfig struct {
	FileserverHits atomic.Int32
	JWTSecret string
	Platform string
	PolkaWebhookSecret string
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

func LoadEnv() (models.EnvVars, error) {
	err := godotenv.Load(".env")
	if err != nil {
		errorMessage := fmt.Errorf("error loading env file: %v", err)
		return models.EnvVars{}, errorMessage
	}

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		errorMessage := fmt.Errorf("DB_URL is not set")
		return models.EnvVars{}, errorMessage
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		errorMessage := fmt.Errorf("PLATFORM is not set")
		return models.EnvVars{}, errorMessage
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		errorMessage := fmt.Errorf("JWT_SECRET is not set")
		return models.EnvVars{}, errorMessage
	}

	polkaWebhookSecret := os.Getenv("POLKA_KEY")
	if polkaWebhookSecret == "" {
		errorMessage := fmt.Errorf("POLKA_KEY is not set")
		return models.EnvVars{}, errorMessage
	}

	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, constants.DEFAULT_PORT)
		port = constants.DEFAULT_PORT
	}

	return models.EnvVars{
		DbUrl: dbUrl,
		JWTSecret: jwtSecret,
		Platform: platform,
		PolkaWebhookSecret: polkaWebhookSecret,
		Port: port,
	}, nil
}
