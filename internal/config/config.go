package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
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

	missingEnvVariables := make([]string, 0)

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		missingEnvVariables = append(missingEnvVariables, "DB_URL")
	}

	platform := os.Getenv("PLATFORM")
	if platform == "" {
		missingEnvVariables = append(missingEnvVariables, "PLATFORM")
	}

	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		missingEnvVariables = append(missingEnvVariables, "JWT_SECRET")
	}

	polkaWebhookSecret := os.Getenv("POLKA_KEY")
	if polkaWebhookSecret == "" {
		missingEnvVariables = append(missingEnvVariables, "POLKA_KEY")
	}

	port, err := getServerPort()
	if err != nil {
		fmt.Printf("error getting server port: %v\nfalling back to default port: %d\n", err, constants.DEFAULT_PORT)
		port = constants.DEFAULT_PORT
	}

	if len(missingEnvVariables) > 0 {
		errorMessage := fmt.Errorf("missing environment variables: %v", strings.Join(missingEnvVariables, ", "))
		return models.EnvVars{}, errorMessage
	}
	
	return models.EnvVars{
		DbUrl: dbUrl,
		JWTSecret: jwtSecret,
		Platform: platform,
		PolkaWebhookSecret: polkaWebhookSecret,
		Port: port,
	}, nil
}
