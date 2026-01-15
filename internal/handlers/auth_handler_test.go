package handlers

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestHandleLogin(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user to login with
		newUser, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request body
		body, err := json.Marshal(models.LoginResource{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Handle the request
		handlers.HandleLogin().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		// Check response
	  var response models.LoginResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &response)
		require.NoError(t, err)

		require.Equal(t, newUser.ID, response.ID)
		require.Equal(t, newUser.Email, response.Email)
		require.Equal(t, newUser.IsChirpyRed, response.IsChirpyRed)
		require.True(t, newUser.CreatedAt.Equal(response.CreatedAt))
		require.True(t, newUser.UpdatedAt.Equal(response.UpdatedAt))
		require.NotEqual(t, response.Token, "")
		require.NotEqual(t, response.RefreshToken, "")

		return nil
	})
}

func TestHandleLogin_Returns_BadRequest_When_Payload_Malformed(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request body
		body := bytes.NewReader([]byte("bleee"))

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", body)
		req.Header.Set("Content-Type", "application/json")

		// Handle the request
		handlers.HandleLogin().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"),)
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}

func TestHandleLogin_Returns_BadRequest_When_Payload_Is_Invalid(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request body
		body, err := json.Marshal(map[string]string{
			"foo": "bar",
			"baz": "qux",
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Handle the request
		handlers.HandleLogin().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"),)
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleLogin_Returns_UnauthorizedError_When_Password_Is_Incorrect(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user to login with
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request body
		body, err := json.Marshal(models.LoginResource{
			Email: testutils.TEST_EMAIL,
			Password: "incorrect password",
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Handle the request
		handlers.HandleLogin().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"),)
		require.Equal(t, "invalid credentials", recorder.Body.String())

		return nil
	})
}

func TestHandleLogin_Returns_NotFoundError_When_Invalid_Email(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user to login with
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request body
		body, err := json.Marshal(models.LoginResource{
			Email: "some-invalid@email.co.uk",
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/login", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		// Handle the request
		handlers.HandleLogin().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"),)
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRefresh(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.TokenRefreshResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.NotEqual(t, responseBody.Token, "")
		require.NotEqual(t, responseBody.Token, loginOutput.Token)

		return nil
	})
}

func TestHandleTokenRefresh_Returns_UnauthorizedError_When_Authorization_Header_Is_Missing(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request without the authorization header
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "no authorization header provided")

		return nil
	})
}

func TestHandleTokenRefresh_Returns_UnauthorizedError_When_Authorization_Header_Is_In_The_Wrong_Format(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Add the authorization header with the wrong format
		req.Header.Set("Authorization", "this-is-a-token")

		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "invalid authorization header format")

		return nil
	})
}

func TestHandleTokenRefresh_Returns_UnauthorizedError_When_Authorization_Token_Missing(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer some-random-token")

		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token not found", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRefresh_Returns_UnauthorizedError_When_Authorization_Token_Expired(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		now := time.Now()
		_, err = q.ExpireRefreshToken(testHelper.Ctx, database.ExpireRefreshTokenParams{
			ExpiresAt: now,
			UpdatedAt: now,
			Token: loginOutput.RefreshToken,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token expired", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRefresh_Returns_UnauthorizedError_When_Authorization_Token_Revoked(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		now := time.Now()
		_, err = q.RevokeRefreshToken(testHelper.Ctx, database.RevokeRefreshTokenParams{
			Token: loginOutput.RefreshToken,
			RevokedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: now,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/refresh", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRefresh().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token revoked", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRevoke(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/revoke", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRevoke().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNoContent, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))

		revokedRefreshToken, err := q.FindRefreshToken(testHelper.Ctx, loginOutput.RefreshToken)
		require.NoError(t, err)
		require.True(t, revokedRefreshToken.RevokedAt.Valid)
		require.True(t, revokedRefreshToken.RevokedAt.Time.Before(time.Now()))

		return nil
	})
}

func TestHandleTokenRevoke_Returns_NotFoundError_When_Authorization_Token_Not_Found(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/revoke", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer some-random-token")

		// Handle the request
		handlers.HandleTokenRevoke().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token not found", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRevoke_Returns_UnauthorizedError_When_Authorization_Token_Already_Expired(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		now := time.Now()
		_, err = q.ExpireRefreshToken(testHelper.Ctx, database.ExpireRefreshTokenParams{
			ExpiresAt: now,
			UpdatedAt: now,
			Token: loginOutput.RefreshToken,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/revoke", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRevoke().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token expired", recorder.Body.String())

		return nil
	})
}

func TestHandleTokenRevoke_Returns_UnauthorizedError_When_Authorization_Token_Already_Revoked(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Create a user
		_, err := handlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := handlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email: testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		now := time.Now()
		_, err = q.RevokeRefreshToken(testHelper.Ctx, database.RevokeRefreshTokenParams{
			Token: loginOutput.RefreshToken,
			RevokedAt: sql.NullTime{Time: now, Valid: true},
			UpdatedAt: now,
		})
		require.NoError(t, err)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/revoke", nil)
		
		// Add the correct authorization header
		req.Header.Set("Authorization", "Bearer " + loginOutput.RefreshToken)

		// Handle the request
		handlers.HandleTokenRevoke().ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "refresh token already revoked", recorder.Body.String())

		return nil
	})
}
