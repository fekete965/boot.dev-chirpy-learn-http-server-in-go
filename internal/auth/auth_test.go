package auth

import (
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
)


func TestHashPassword(t *testing.T) {
	testPassword := "test_password"
	_, err := HashPassword(testPassword)

	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestHashPassword: %v", testPassword, err)
	}

}

func TestCheckPasswordHashWithCorrectPassword(t *testing.T) {
	testPassword := "test_password"
	
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestCheckPasswordHash: %v", testPassword, err)
	}
	
	match, err := CheckPasswordHash(testPassword, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash(%v, %v) returned an error in TestCheckPasswordHash: %v", testPassword, hash,err)
	}

	if !match {
		t.Errorf("CheckPasswordHash(%v, %v) returned %v instead of %v in TestCheckPasswordHash", testPassword, hash, match, true)
	}
}
func TestCheckPasswordHashWithWrongPassword(t *testing.T) {
	testPassword := "test_password"
	wrongPassword := "wrong_password"
	
	hash, err := HashPassword(testPassword)
	if err != nil {
		t.Errorf("HashPassword(%v) returned an error in TestCheckPasswordHash: %v", testPassword, err)
	}
	
	match, err := CheckPasswordHash(wrongPassword, hash)
	if err != nil {
		t.Errorf("CheckPasswordHash(%v, %v) returned an error in TestCheckPasswordHash: %v", wrongPassword, hash,err)
	}

	if match {
		t.Errorf("CheckPasswordHash(%v, %v) returned %v instead of %v in TestCheckPasswordHash", wrongPassword, hash, match, false)
	}
}

func TestMakeJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test_token_secret"
	expiresIn := 3 * time.Hour

	_, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT(%v, %v, %v) returned an error in TestMakeJWT: %v", userID, tokenSecret, expiresIn, err)
	}
}

func TestValidateJWT(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test_token_secret"
	expiresIn := 3 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT(%v, %v, %v) returned an error in TestValidateJWT: %v", userID, tokenSecret, expiresIn, err)
	}

	validatedUserID, err := ValidateJWT(token, tokenSecret)
	if err != nil {
		t.Errorf("ValidatedJWT(%v, %v) returned an error in TestValidateJWT: %v", token, tokenSecret, err)
	}

	if validatedUserID != userID {
		t.Errorf("The user id %v returned by ValidatedJWT(%v, %v) doesn't match the expected user id %v", validatedUserID, token, tokenSecret, userID)
	}
}

func TestValidateJWTWithExpiredToken(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test_token_secret"
	expiresIn := -1 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT(%v, %v, %v) returned an error in TestValidateJWTWithExpiredToken: %v", userID, tokenSecret, expiresIn, err)
	}

	_, err = ValidateJWT(token, tokenSecret)
	if err == nil {
		t.Errorf("ValidatedJWT(%v, %v) should have returned an error in TestValidateJWTWithExpiredToken: %v", token, tokenSecret, err)
	}

	if !strings.Contains(err.Error(), "token is expired") {
		t.Errorf("ValidatedJWT(%v, %v) should have returned \"token is expired\" error instead of %v", token, tokenSecret, err.Error())
	}
}

func TestValidateJWTWithWrongSecret(t *testing.T) {
	userID := uuid.New()
	tokenSecret := "test_token_secret"
	wrongTokenSecret := "wrong_token_secret"
	expiresIn := 3 * time.Hour

	token, err := MakeJWT(userID, tokenSecret, expiresIn)
	if err != nil {
		t.Errorf("MakeJWT(%v, %v, %v) returned an error in TestValidateJWTWithExpiredToken: %v", userID, tokenSecret, expiresIn, err)
	}

	_, err = ValidateJWT(token, wrongTokenSecret)
	if err == nil {
		t.Errorf("ValidatedJWT(%v, %v) should have returned an error in TestValidateJWTWithWrongSecret: %v", token, wrongTokenSecret, err)

func TestGetBearerToken(t *testing.T) {
	testToken := "test_token"

	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	request.Header.Set("Authorization", "Bearer " + testToken)

	bearerToken, err := GetBearerToken(request)
	if err != nil {
		t.Errorf("GetBearerToken(...) returned an error: %v", err)
	}

	if bearerToken != testToken {
		t.Errorf("GetBearerToken(...) returned %v instead of %v", bearerToken, testToken)
	}
}

func TestGetBearerTokenWithoutAuthorizationHeader(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}
	
	// Request is missing the Authorization header
	_, err = GetBearerToken(request)
	if err == nil {
		t.Error("GetBearerToken(...) should have returned and error stating that no authorization header was provided")
	}
}

func TestGetBearerTokenWithInvalidAuthorizationHeaderFormat(t *testing.T) {
	testToken := "test_token"

	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	// Add Authorization header without the correct prefix
	request.Header.Set("Authorization", testToken)

	_, err = GetBearerToken(request)
	if err == nil {
		t.Error("GetBearerToken(...) should have returned and error stating that the authorization header format is invalid")
	}
}
