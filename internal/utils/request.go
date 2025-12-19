package utils

import (
	"encoding/json"
	"log"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)

func GetQueryParam(r *http.Request, paramName string, defaultValue *string) *string {
	param := r.URL.Query().Get(paramName)
	if param == "" {
		return defaultValue
	}

	return &param
}

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

	return data, nil
}

func GetBearerToken(r *http.Request) (string, error) {
	bearerToken, err := auth.GetBearerToken(r)
	if err != nil {
		return "", service_errors.NewUnauthorizedError("error during token retrieval")
	}

	return bearerToken, nil
}

