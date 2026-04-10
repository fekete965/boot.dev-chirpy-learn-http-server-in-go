package utils

import (
	"errors"
	"net/http"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/stretchr/testify/require"
)

func TestServiceErrorToApiResponseWithGenericError(t *testing.T) {
	someError := errors.New("some error")
	statusCode, _ := ServiceErrorToApiResponse(someError)

	require.Equal(t, http.StatusInternalServerError, statusCode)
	require.Equal(t, someError.Error(), "some error")
}

func TestServiceErrorToApiResponseWithDecodeError(t *testing.T) {
	decodeError := models.NewDecodeRequestBodyError(models.NewDecodeRequestBodyErrorInput{
		Code:        "DECODE_ERROR",
		Message:     "something went wrong during decoding",
		FieldErrors: nil,
	})

	statusCode, apiErrorResponse := ServiceErrorToApiResponse(decodeError)

	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Equal(t, "DECODE_ERROR", apiErrorResponse.Error.Code)
	require.Equal(t, "something went wrong during decoding", apiErrorResponse.Error.Message)
	require.Nil(t, apiErrorResponse.Error.Fields)
}

func TestServiceErrorToApiResponseWithDecodeErrorWithTheCorrectFields(t *testing.T) {
	aer := models.NewDecodeRequestBodyError(models.NewDecodeRequestBodyErrorInput{
		Code:    "DECODE_ERROR",
		Message: "something went wrong during resource decoding",
		FieldErrors: map[string][]string{
			"name": {"Name is required"},
			"age":  {"Age must be > 18"},
		},
	})
	statusCode, apiErrorResponse := ServiceErrorToApiResponse(aer)

	require.Equal(t, http.StatusBadRequest, statusCode)
	require.Equal(t, "DECODE_ERROR", apiErrorResponse.Error.Code)
	require.Equal(t, "something went wrong during resource decoding", apiErrorResponse.Error.Message)
	require.NotEqual(t, nil, apiErrorResponse.Error.Fields)

	require.Equal(t, map[string][]string{
		"name": {"Name is required"},
		"age":  {"Age must be > 18"},
	}, apiErrorResponse.Error.Fields)
}

func TestServiceErrorToApiResponseWithServiceError(t *testing.T) {
	type testCases struct {
		Name                 string
		ServiceError         error
		ExpectedStatusCode   int
		ExpectedErrorMessage string
		ExpectedErrorCode    string
	}

	tests := []testCases{
		{
			Name:                 "found error case",
			ServiceError:         service_errors.NewNotFoundError("user cannot be found"),
			ExpectedStatusCode:   http.StatusNotFound,
			ExpectedErrorMessage: "user cannot be found",
			ExpectedErrorCode:    "NOT_FOUND",
		},
		{
			Name:                 "conflict error case",
			ServiceError:         service_errors.NewConflictError("conflict error"),
			ExpectedStatusCode:   http.StatusConflict,
			ExpectedErrorMessage: "conflict error",
			ExpectedErrorCode:    "CONFLICT",
		},
		{
			Name:                 "bad request case",
			ServiceError:         service_errors.NewBadRequestError("bad request error"),
			ExpectedStatusCode:   http.StatusBadRequest,
			ExpectedErrorMessage: "bad request error",
			ExpectedErrorCode:    "BAD_REQUEST",
		},
		{
			Name:                 "unauthorized error case",
			ServiceError:         service_errors.NewUnauthorizedError("unauthorized error"),
			ExpectedStatusCode:   http.StatusUnauthorized,
			ExpectedErrorMessage: "unauthorized error",
			ExpectedErrorCode:    "UNAUTHORIZED",
		},
		{
			Name:                 "forbidden error case",
			ServiceError:         service_errors.NewForbiddenError("forbidden error"),
			ExpectedStatusCode:   http.StatusForbidden,
			ExpectedErrorMessage: "forbidden error",
			ExpectedErrorCode:    "FORBIDDEN",
		},
		{
			Name:                 "internal server error case",
			ServiceError:         service_errors.NewInternalServerError("internal server error"),
			ExpectedStatusCode:   http.StatusInternalServerError,
			ExpectedErrorMessage: "internal server error",
			ExpectedErrorCode:    "INTERNAL_SERVER_ERROR",
		},
	}

	for _, test := range tests {
		t.Run(test.Name, func(t *testing.T) {
			statusCode, apiErrorResponse := ServiceErrorToApiResponse(test.ServiceError)

			require.Equal(t, test.ExpectedStatusCode, statusCode)
			require.Equal(t, test.ExpectedErrorMessage, apiErrorResponse.Error.Message)
			require.Equal(t, test.ExpectedErrorCode, apiErrorResponse.Error.Code)
		})
	}
}
