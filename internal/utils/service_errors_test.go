package utils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)

func TestServiceErrorToRequestErrorWithGenericError(t *testing.T) {
	genericError := errors.New("some generic error")
	statusCode, errorMessage := ServiceErrorToRequestError(genericError)

	expectedStatusCode := http.StatusInternalServerError
	if statusCode != expectedStatusCode {
		t.Errorf("Expected %d status code, but received %d", expectedStatusCode, statusCode)
	}
	
	expectedErrorMessage := genericError.Error()
	if errorMessage != expectedErrorMessage {
		t.Errorf("Expected %s error message, but received %s", expectedErrorMessage, errorMessage)
	}
}

func TestServiceErrorToRequestErrorWithServiceError(t *testing.T) {
	serviceError := service_errors.NewNotFoundError("user cannot be found")
	statusCode, errorMessage := ServiceErrorToRequestError(serviceError)

	expectedStatusCode := http.StatusNotFound
	if statusCode != expectedStatusCode {
		t.Errorf("Expected %d status code, but received %d", expectedStatusCode, statusCode)
	}

	expectedErrorMessage := serviceError.Error()
	if errorMessage != expectedErrorMessage {
		t.Errorf("Expected %s error message, but received %s", expectedErrorMessage, errorMessage)
	}
}
