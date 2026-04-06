package utils

import (
	"fmt"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

func CleanChirp(textToClean string, profaneWords []string) string {
	chunks := strings.Split(textToClean, " ")

	for cIndex, c := range chunks {
		for _, word := range profaneWords {
			if strings.EqualFold(c, word) {
				chunks[cIndex] = strings.Repeat("*", 4)
			}
		}
	}

	return strings.Join(chunks, " ")
}

func SafeParseUUID(str string) (uuid.UUID, error) {
	err := uuid.Validate(str)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("invalid UUID: %v", err)
	}

	return uuid.MustParse(str), nil
}

func ValidateChirpLength(chirp string, maxLength int) error {
	if len(chirp) > maxLength {
		return fmt.Errorf("chirp is too long")
	}

	return nil
}

type FieldErrorMessageGetter = func(fe validator.FieldError) string

var DefaultFieldErrorMessageGetter FieldErrorMessageGetter = func(fe validator.FieldError) string {
	return fmt.Sprintf("'%s' is invalid", cases.Title(language.English).String(fe.Field()))
}

var FieldErrorValidationMessageGetters = map[string]FieldErrorMessageGetter{
	"email": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be a valid email address",
			cases.Title(language.English).String(fe.Field()),
		)
	},
	"max": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be less than %s characters",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
		)
	},
	"min": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be at least %s characters long",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
		)
	},
	"lte": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be ≤ %s (got: %v)",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
			fe.Value(),
		)
	},
	"gte": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be ≥ %s (got: %v)",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
			fe.Value(),
		)
	},
	"lt": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be < %s (got: %v)",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
			fe.Value(),
		)
	},
	"gt": func(fe validator.FieldError) string {
		return fmt.Sprintf(
			"'%s' must be > %s (got: %v)",
			cases.Title(language.English).String(fe.Field()),
			fe.Param(),
			fe.Value(),
		)
	},
	"required": func(fe validator.FieldError) string {
		return fmt.Sprintf("'%s' is required",
			cases.Title(language.English).String(fe.Field()),
		)
	},
}
