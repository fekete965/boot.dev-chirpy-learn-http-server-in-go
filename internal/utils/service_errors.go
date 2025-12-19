package utils

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)

func ServiceErrorToRequestError(err error) (statusCode int, errorMessage string, ) {
	switch e := err.(type) {
		case *service_errors.BadRequestError:
			return e.Code, e.Message
		case *service_errors.NotFoundError:
			return e.Code, e.Message
		case *service_errors.ConflictError:
			return e.Code, e.Message
		case *service_errors.UnauthorizedError:
			return e.Code, e.Message
		case *service_errors.ForbiddenError:
			return e.Code, e.Message
		case *service_errors.InternalServerError:
			return e.Code, e.Message
		default:
			return http.StatusInternalServerError, err.Error()
	}
}
