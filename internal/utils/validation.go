package utils

import (
	"fmt"
	"strings"

	"github.com/google/uuid"
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
