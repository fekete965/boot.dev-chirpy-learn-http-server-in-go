package utils

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"reflect"
	"strings"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
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

func initValidator() *validator.Validate {
	validate := validator.New(validator.WithRequiredStructEnabled())

	validate.RegisterTagNameFunc(func(fld reflect.StructField) string {
		name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
		if name == "-" {
			return ""
		}
	
		return name
	})

	return validate
}
var validate *validator.Validate = initValidator()

func DecodeRequestBody[T any](r *http.Request) (T, models.DecodeRequestBodyError) {
	var data T
	var zero T

	decoder := json.NewDecoder(r.Body)
	defer r.Body.Close()

	err := decoder.Decode(&data)
	if err != nil {
		errorMessage := "error decoding request body"
		log.Printf("%s: %v", errorMessage, err)

		return zero, models.DecodeRequestBodyError{
			Code: "DECODE_ERROR",
			Message: errorMessage,
			FieldErrors: nil,
		}
	}

	err = validate.Struct(data)
	if err != nil {
		var validateErrs validator.ValidationErrors
		fieldErrors := make(map[string][]string)
		
		if errors.As(err, &validateErrs) {
			for _, valErr := range validateErrs {
				msgGetter, ok := FieldErrorValidationMessageGetters[valErr.Tag()]
				if !ok {
					msgGetter = DefaultFieldErrorMessageGetter
				}

				msg := msgGetter(valErr)
				fieldErrors[valErr.Field()] = append(fieldErrors[valErr.Field()], msg)
			}
		}

		errorMessage := "validation error"
		log.Print(errorMessage)

		return zero, models.DecodeRequestBodyError{
			Code: "VALIDATION_ERROR",
			Message: errorMessage,
			FieldErrors: fieldErrors,
		}
	}

	return data, models.DecodeRequestBodyError{}
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
