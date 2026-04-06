package utils

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/google/uuid"
)

type testPayload struct {
	Message string `json:"message"`
}

var (
	testParamName         string = "test"
	testParamValue        string = "test-value"
	testDefaultParamValue string = "default-test-value"
	testPayloadValue      string = "Hello, there!"
	testToken             string = "test-token"
	testJwtSecret         string = "test-jwt-secret"
)

func getTestPayload(t *testing.T) io.Reader {
	testPayloadValue := testPayload{Message: testPayloadValue}
	stringifiedPayload, err := json.Marshal(testPayloadValue)
	if err != nil {
		t.Errorf("json.Marshal(%v) returned an error: %v", testPayloadValue, err)
	}

	return bytes.NewBuffer(stringifiedPayload)
}

func getValidBearerToken(t *testing.T, userID uuid.UUID, expiresIn time.Duration) string {
	token, err := auth.MakeJWT(userID, testJwtSecret, expiresIn)
	if err != nil {
		t.Errorf("auth.MakeJWT(...) returned an error: %v", err)
	}

	return token
}

func TestGetQueryParam(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	queryParam := GetQueryParam(request, testParamName, nil)
	if queryParam != nil {
		t.Errorf("Query param %s should be nil, but got %s", testParamName, *queryParam)
	}

	query := request.URL.Query()
	query.Add(testParamName, testParamValue)

	request.URL.RawQuery = query.Encode()

	queryParam = GetQueryParam(request, testParamName, nil)
	if queryParam == nil {
		t.Errorf("Query param %s should not be nil", testParamName)
	}

	if queryParam != nil && *queryParam != testParamValue {
		t.Errorf("Expected %s, but got %s", testParamValue, *queryParam)
	}
}

func TestGetQueryParamWithDefaultValue(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\",nil) returned an error: %v", err)
	}

	queryParam := GetQueryParam(request, testParamName, &testDefaultParamValue)
	if queryParam == nil {
		t.Errorf("Query param %s should not be nil", testParamName)
	}

	if queryParam != nil && *queryParam != testDefaultParamValue {
		t.Errorf("Expected %s, but got %s", testDefaultParamValue, *queryParam)
	}
}

func TestGetQueryParamWithEmptyValue(t *testing.T) {
	request, err := http.NewRequest("GET", "/?test=", nil)
	if err != nil {
		t.Errorf("NewRequest returned an error: %v", err)
	}

	queryParam := GetQueryParam(request, testParamName, &testDefaultParamValue)
	if queryParam == nil {
		t.Errorf("Query param should not be nil")
	}

	if queryParam != nil && *queryParam != testDefaultParamValue {
		t.Errorf("Expected default value %s, got %s", testDefaultParamValue, *queryParam)
	}
}

func TestDecodeRequestBody(t *testing.T) {
	payload := getTestPayload(t)

	request, err := http.NewRequest("POST", "/", payload)
	if err != nil {
		t.Errorf("NewRequest(\"POST\", \"/\", payload) returned an error: %v", err)
	}

	data, err := DecodeRequestBody[testPayload](request)
	if err != nil {
		t.Errorf("DecodeRequestBody[testPayload](request.Body) returned an error: %v", err)
	}

	if data.Message != testPayloadValue {
		t.Errorf("Expected %s, but got %s", testPayloadValue, data.Message)
	}
}

func TestDecodeRequestBodyWithInvalidPayload(t *testing.T) {
	invalidPayload := bytes.NewBuffer([]byte("invalid payload"))

	request, err := http.NewRequest("POST", "/", invalidPayload)
	if err != nil {
		t.Errorf("NewRequest(\"POST\", \"/\", invalidPayload) returned an error: %v", err)
	}

	_, err = DecodeRequestBody[testPayload](request)
	if err == nil {
		t.Errorf("DecodeRequestBody[testPayload](request.Body) should have returned an error")
	}
}

func TestDecodeRequestBodyWithEmptyBody(t *testing.T) {
	request, err := http.NewRequest("POST", "/", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Errorf("NewRequest returned an error: %v", err)
	}

	_, err = DecodeRequestBody[testPayload](request)
	if err == nil {
		t.Errorf("DecodeRequestBody should have returned an error for empty body")
	}
}

func TestGetBearerToken(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+testToken)

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

	_, err = GetBearerToken(request)
	if err == nil {
		t.Errorf("GetBearerToken(...) should be an error stating that no authorization header was provided")
	}

	expectedErrorPart := "no authorization header provided"
	if !strings.Contains(err.Error(), expectedErrorPart) {
		t.Errorf("GetBearerToken(...) returned an unexpected error: %v", err)
	}
}

func TestGetBearerTokenWithInvalidAuthorizationHeaderFormat(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	request.Header.Set("Authorization", testToken)

	_, err = GetBearerToken(request)
	if err == nil {
		t.Errorf("GetBearerToken(...) should be an error stating that the authorization header format is invalid")
	}

	expectedErrorPart := "invalid authorization header format"
	if !strings.Contains(err.Error(), expectedErrorPart) {
		t.Errorf("GetBearerToken(...) returned an unexpected error: %v", err)
	}
}

func TestGetAuthenticatedUserID(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	userID := uuid.New()
	validBearerToken := getValidBearerToken(t, userID, 3*time.Hour)
	request.Header.Set("Authorization", "Bearer "+validBearerToken)

	authenticatedUserID, token, err := GetAuthenticatedUserID(request, testJwtSecret)
	if err != nil {
		t.Errorf("GetAuthenticatedUserID(...) returned an error: %v", err)
	}

	if authenticatedUserID != userID {
		t.Errorf("GetAuthenticatedUserID(...) returned %v instead of %v", authenticatedUserID, userID)
	}

	if token != validBearerToken {
		t.Errorf("GetAuthenticatedUserID(...) returned %v instead of %v", token, validBearerToken)
	}
}

func TestGetAuthenticatedUserIDWithInvalidBearerToken(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	request.Header.Set("Authorization", "Bearer "+"invalid-token")

	_, _, err = GetAuthenticatedUserID(request, testJwtSecret)
	if err == nil {
		t.Errorf("GetAuthenticatedUserID(...) should have returned an error stating that the bearer token is invalid")
	}
}

func TestGetAuthenticatedUserIDWithExpiredBearerToken(t *testing.T) {
	request, err := http.NewRequest("GET", "/", nil)
	if err != nil {
		t.Errorf("NewRequest(\"GET\", \"/\", nil) returned an error: %v", err)
	}

	userID := uuid.New()
	validBearerToken := getValidBearerToken(t, userID, -1*time.Hour)
	request.Header.Set("Authorization", "Bearer "+validBearerToken)

	_, _, err = GetAuthenticatedUserID(request, testJwtSecret)
	if err == nil {
		t.Errorf("GetAuthenticatedUserID(...) should have returned an error stating that the bearer token is expired")
	}
}
