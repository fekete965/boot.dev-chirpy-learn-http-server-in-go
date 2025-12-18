package utils

import (
	"errors"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)

func ServiceErrorToRequestError(err error) (statusCode int, errorMessage string, ) {
	var serviceErr *service_errors.ServiceError
	if errors.As(err, &serviceErr) {
		return serviceErr.Code, serviceErr.Message
	}

	return http.StatusInternalServerError, err.Error()
}
