package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
)

var (
	dbUrl string = "DB_URL=postgres://user:password@localhost/dbname?sslmode=disable"
	platform string = "PLATFORM=dev"
	polkaKey string = "POLKA_KEY=polka-webhook-secret"
	jwtSecret string = "JWT_SECRET=jwt-secret"
	port string = "PORT=3054"
)

func clearEnvVars() {
	os.Unsetenv("DB_URL")
	os.Unsetenv("PLATFORM")
	os.Unsetenv("POLKA_KEY")
	os.Unsetenv("JWT_SECRET")
	os.Unsetenv("PORT")
}

func setupTest(t *testing.T, fileContent []string) (cwd string) {
	clearEnvVars()
	
	// Create a temporary directory
	tmpDir := t.TempDir()
	
	// Create a temporary .env file in the temporary directory
	envFile := filepath.Join(tmpDir, ".env")
	envFileContent := strings.Join(fileContent, "\n")
	if err := os.WriteFile(envFile, []byte(envFileContent), 0644); err != nil {
		t.Fatalf("failed to create temporary .env file: %v", err)
	}

	// Get the current working directory
	cwd, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get the current working directory: %v", err)
	}

	// Navigate into the temporary directory so "godotenv.Load()"
	// could find the temporary .env file
	os.Chdir(tmpDir)	

	return cwd
}

func cleanupTest(t *testing.T, cwd string) {
	if err := os.Chdir(cwd); err != nil {
		t.Fatalf("failed to change back to the current working directory: %v", err)
	}
}

func TestLoadEnvSuccess(t *testing.T) {
	fileContent := []string{dbUrl, platform, polkaKey, jwtSecret, port}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	envVars, err := LoadEnv()
	if err != nil {
		t.Errorf("LoadEnv() returned an error: %v", err)
	}

	if envVars.DbUrl != "postgres://user:password@localhost/dbname?sslmode=disable" {
		t.Errorf("DbUrl is not set to the expected value: %v", envVars.DbUrl)
	}
	
	if envVars.Platform != "dev" {
		t.Errorf("Platform is not set to the expected value: %v", envVars.Platform)
	}
	
	if envVars.PolkaWebhookSecret != "polka-webhook-secret" {
		t.Errorf("PolkaWebhookSecret is not set to the expected value: %v", envVars.PolkaWebhookSecret)
	}
	
	if envVars.JWTSecret != "jwt-secret" {
		t.Errorf("JWTSecret is not set to the expected value: %v", envVars.JWTSecret)
	}
	
	if envVars.Port != 3054 {
		t.Errorf("Port is not set to the expected value: %v", envVars.Port)
	}
}

func TestLoadEnvMissing_PORT(t *testing.T) {
	fileContent := []string{dbUrl, platform, polkaKey, jwtSecret}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	envVars, err := LoadEnv()
	if err != nil {
		t.Errorf("LoadEnv() returned an error: %v", err)
	}

	if envVars.Port != 8080 {
		t.Errorf("Expected port to be set to the fallback value of %v instead received: %v", constants.DEFAULT_PORT, envVars.Port)
	}
}

func TestLoadEnvMissing_JWT_SECRET(t *testing.T) {
	fileContent := []string{dbUrl, platform, polkaKey, port}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	_, err := LoadEnv()
	if err == nil {
		t.Errorf("LoadEnv() should have returned an error stating that the JWT_SECRET is missing")
	}

	if err.Error() != "missing environment variables: JWT_SECRET" {
		t.Errorf("LoadEnv() returned an unexpected error: %v", err.Error())
	}
}

func TestLoadEnvMissing_POLKA_KEY(t *testing.T) {
	fileContent := []string{dbUrl, platform, jwtSecret, port}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	_, err := LoadEnv()
	if err == nil {
		t.Errorf("LoadEnv() should have returned an error stating that the POLKA_KEY is missing")
	}

	if err.Error() != "missing environment variables: POLKA_KEY" {
		t.Errorf("LoadEnv() returned an unexpected error: %v", err.Error())
	}
}

func TestLoadEnvMissing_PLATFORM(t *testing.T) {
	fileContent := []string{dbUrl, polkaKey, jwtSecret, port}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	_, err := LoadEnv()
	if err == nil {
		t.Errorf("LoadEnv() should have returned an error stating that the PLATFORM is missing")
	}

	if err.Error() != "missing environment variables: PLATFORM" {
		t.Errorf("LoadEnv() returned an unexpected error: %v", err.Error())
	}
}

func TestLoadEnvMissing_DB_URL(t *testing.T) {
	fileContent := []string{platform, polkaKey, jwtSecret, port}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)
	
	_, err := LoadEnv()
	if err == nil {
		t.Errorf("LoadEnv() should have returned an error stating that the DB_URL is missing")
	}

	if err.Error() != "missing environment variables: DB_URL" {
		t.Errorf("LoadEnv() returned an unexpected error: %v", err.Error())
	}
}

func TestLoadEnvMultipleMissingValues(t *testing.T) {
	fileContent := []string{}
	cwd := setupTest(t, fileContent)
	defer cleanupTest(t, cwd)

	_, err := LoadEnv()
	if err == nil {
		t.Errorf("LoadEnv() should have returned an error stating that multiple environment variables are missing")
	}

	if err.Error() != "missing environment variables: DB_URL, PLATFORM, JWT_SECRET, POLKA_KEY" {
		t.Errorf("LoadEnv() returned an unexpected error: %v", err.Error())
	}
}
