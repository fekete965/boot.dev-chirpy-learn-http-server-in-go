package utils

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)

func ServiceErrorToApiResponse(err error) (int, models.ApiErrorResponse) {
	var drbe *models.DecodeRequestBodyError
	if errors.As(err, &drbe) {
		return http.StatusBadRequest, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    drbe.Code,
				Message: drbe.Message,
				Fields:  drbe.FieldErrors,
			},
		}
	}
	var bre *service_errors.BadRequestError
	if errors.As(err, &bre) {
		return bre.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "BAD_REQUEST",
				Message: bre.Message,
				Fields:  nil,
			},
		}
	}
	var nfe *service_errors.NotFoundError
	if errors.As(err, &nfe) {
		return nfe.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "NOT_FOUND",
				Message: nfe.Message,
				Fields:  nil,
			},
		}
	}
	var cfe *service_errors.ConflictError
	if errors.As(err, &cfe) {
		return cfe.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "CONFLICT",
				Message: cfe.Message,
				Fields:  nil,
			},
		}
	}
	var ue *service_errors.UnauthorizedError
	if errors.As(err, &ue) {
		return ue.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "UNAUTHORIZED",
				Message: ue.Message,
				Fields:  nil,
			},
		}
	}
	var fre *service_errors.ForbiddenError
	if errors.As(err, &fre) {
		return fre.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "FORBIDDEN",
				Message: fre.Message,
				Fields:  nil,
			},
		}
	}
	var ise *service_errors.InternalServerError
	if errors.As(err, &ise) {
		return ise.Code, models.ApiErrorResponse{
			Error: models.ApiError{
				Code:    "INTERNAL_SERVER_ERROR",
				Message: ise.Message,
				Fields:  nil,
			},
		}
	}

	// TODO: Log instead of this
	errMsg := "Unexpected error"
	fmt.Printf("%s: %v", errMsg, err.Error())
	return http.StatusInternalServerError, models.ApiErrorResponse{
		Error: models.ApiError{
			Code:    "INTERNAL_SERVER_ERROR",
			Message: errMsg,
			Fields:  nil,
		},
	}
}
