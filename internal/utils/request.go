package utils

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

func GetQueryParam(r *http.Request, paramName string, defaultValue *string) *string {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return defaultValue
	}

	return &param
}

var validate *validator.Validate = validator.New(validator.WithRequiredStructEnabled())

func DecodeRequestBody[T any](r *http.Request) (T, error) {
	var data T

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&data)
	if err != nil {
		errorMessage := "error decoding request body"
		log.Printf("%s: %v", errorMessage, err)
		return data, service_errors.NewBadRequestError(errorMessage)
	}

	validate = validator.New(validator.WithRequiredStructEnabled())
	err = validate.Struct(data)
	if err != nil {
		errorMessage := "validation error"
		log.Print(errorMessage)
		return data, service_errors.NewBadRequestError(errorMessage)
	}

	return data, nil
}

func GetBearerToken(r *http.Request) (string, error) {
	bearerToken, err := auth.GetBearerToken(r)
	if err != nil {
		errorMessage := fmt.Sprintf("error during token retrieval: %v", err)
		return "", service_errors.NewUnauthorizedError(errorMessage)
	}

	return bearerToken, nil
}

func GetAuthenticatedUserID(r *http.Request, jwtSecret string) (userID uuid.UUID, token string, err error) {
	bearerToken, err := GetBearerToken(r)
	if err != nil {
		errorMessage := fmt.Sprintf("error during token retrieval: %v", err)
		return uuid.UUID{}, "", service_errors.NewUnauthorizedError(errorMessage)
	}
	
	userID, err = auth.ValidateJWT(bearerToken, jwtSecret)
	if err != nil {
		errorMessage := fmt.Sprintf("error during validation: %v", err)
		return uuid.UUID{}, "", service_errors.NewUnauthorizedError(errorMessage)
	}

	return userID, bearerToken, nil
}
