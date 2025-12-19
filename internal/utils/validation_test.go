package utils

import (
	"testing"

	"github.com/google/uuid"
)

func TestCleanChirpCleansTextCorrectly(t *testing.T) {
	profaneWords := []string{"freakin"}
	result := CleanChirp("This is a freakin great day!", profaneWords)
	expected := "This is a **** great day!"

	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestCleanChirpCleansShouldNotCleanTheText(t *testing.T) {
	profaneWords := []string{}
	result := CleanChirp("This is a freakin great day!", profaneWords)
	expected := "This is a freakin great day!"
	
	if result != expected {
		t.Errorf("Expected %s, but got %s", expected, result)
	}
}

func TestSafeParseUUIDReturnValidUUID(t *testing.T) {
	uuidString := "123e4567-e89b-12d3-a456-426614174000"

	parsedUUID, err := SafeParseUUID(uuidString)
	if err != nil {
		t.Errorf("SafeParseUUID(%v) returned an error: %v", uuidString, err)
	}

	if parsedUUID.String() != uuidString {
		t.Errorf("SafeParseUUID(%v) returned %v instead of %v", uuidString, parsedUUID.String(), uuidString)
	}
}

func TestSafeParseUUIDReturnsErrorForInvalidUUID(t *testing.T) {
	invalidUUIDString := "invalid-uuid"

	parsedUUID, err := SafeParseUUID(invalidUUIDString)
	if err == nil {
		t.Errorf("Expected SafeParseUUID(%v) to returned an error", invalidUUIDString)
	}

	if parsedUUID != (uuid.UUID{}) {
		t.Errorf("Expected SafeParseUUID(%v) to returned an empty UUID", invalidUUIDString)
	}
}

func TestValidateChirpLengthWithValidChirp(t *testing.T) {
	validChirp := "this is a valid chirp"
	err := ValidateChirpLength(validChirp, 140)
	if err != nil {
		t.Errorf("ValidateChirpLength(%v) returned an error: %v", validChirp, err)
	}
}

func TestValidateChirpLengthWithInvalidChirp(t *testing.T) {
	invalidChirp := "this is a invalid chirp that is too long"
	err := ValidateChirpLength(invalidChirp, 10)
	if err == nil {
		t.Errorf("ValidateChirpLength(%v) should have returned an error", invalidChirp)
	}

	if err.Error() != "chirp is too long" {
		t.Errorf("ValidateChirpLength(%v) should have returned \"chirp is too long\" error instead of %v", invalidChirp, err.Error())
	}
}
