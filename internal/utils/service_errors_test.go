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
	type testCases struct {
		ServiceError error
		ExpectedStatusCode int
		ExpectedErrorMessage string
	}
	
	tests := []testCases {
		{
			ServiceError: service_errors.NewNotFoundError("user cannot be found"),
			ExpectedStatusCode: http.StatusNotFound,
			ExpectedErrorMessage: "user cannot be found",
		},
		{
			ServiceError: service_errors.NewConflictError("conflict error"),
			ExpectedStatusCode: http.StatusConflict,
			ExpectedErrorMessage: "conflict error",
		},
		{
			ServiceError: service_errors.NewBadRequestError("bad request error"),
			ExpectedStatusCode: http.StatusBadRequest,
			ExpectedErrorMessage: "bad request error",
		},
		{
			ServiceError: service_errors.NewUnauthorizedError("unauthorized error"),
			ExpectedStatusCode: http.StatusUnauthorized,
			ExpectedErrorMessage: "unauthorized error",
		},
		{
			ServiceError: service_errors.NewForbiddenError("forbidden error"),
			ExpectedStatusCode: http.StatusForbidden,
			ExpectedErrorMessage: "forbidden error",
		},
		{
			ServiceError: service_errors.NewInternalServerError("internal server error"),
			ExpectedStatusCode: http.StatusInternalServerError,
			ExpectedErrorMessage: "internal server error",
		},
	}

	for _, test := range tests {
		statusCode, errorMessage := ServiceErrorToRequestError(test.ServiceError)

		if statusCode != test.ExpectedStatusCode {
			t.Errorf("Expected %d status code, but received %d", test.ExpectedStatusCode, statusCode)
		}
	
		if errorMessage != test.ExpectedErrorMessage {
			t.Errorf("Expected %s error message, but received %s", test.ExpectedErrorMessage, errorMessage)
		}
	}
}
