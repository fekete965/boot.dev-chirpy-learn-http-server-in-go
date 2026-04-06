package router

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestHandleCreateUser(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		body, err := json.Marshal(models.CreateUserResource{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusCreated, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.CreateUserResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, testutils.TEST_EMAIL, responseBody.Email)
		require.Equal(t, false, responseBody.IsChirpyRed)
		require.False(t, responseBody.CreatedAt.IsZero())
		require.False(t, responseBody.UpdatedAt.IsZero())

		return nil
	})
}

func TestHandleCreateUser_Returns_BadRequest_When_Payload_Is_Malformed(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", strings.NewReader("malformed body"))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateUser_Returns_BadRequest_When_Email_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		body, err := json.Marshal(models.CreateUserResource{
			Email:    "invalid email",
			Password: "duckling",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateUser_Returns_ConflictError_When_Email_Already_In_Use(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		body, err := json.Marshal(models.CreateUserResource{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusCreated, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		// Try to register the same user again
		recorder2 := httptest.NewRecorder()
		req2 := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req2.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder2, req2)
		require.Equal(t, http.StatusConflict, recorder2.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder2.Header().Get("Content-Type"))
		require.Equal(t, "email already exists", recorder2.Body.String())

		return nil
	})
}

func TestHandleCreateUser_Returns_BadRequest_When_Email_Is_Empty(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		body, err := json.Marshal(models.CreateUserResource{
			Email:    "",
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateUser_Returns_BadRequest_When_Password_Is_Empty(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		body, err := json.Marshal(models.CreateUserResource{
			Email:    testutils.TEST_EMAIL,
			Password: "",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.UpdateUserResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, "new@email.co.uk", responseBody.Email)
		require.Equal(t, false, responseBody.IsChirpyRed)
		require.False(t, responseBody.CreatedAt.IsZero())
		require.False(t, responseBody.UpdatedAt.IsZero())

		return nil
	})
}

func TestHandleUpdateUser_Returns_BadRequest_When_Payload_Is_Malformed(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", strings.NewReader("malformed body"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser_Returns_BadRequest_When_Email_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "invalid email",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser_Returns_ConflictError_When_Email_Already_In_Use(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Create another user
		anotherUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    anotherUser.Email,
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusConflict, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "email already exists", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser_Returns_NotFoundError_When_User_Not_Found(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		_, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// For some reason the user got deleted
		err = q.DeleteAllUsers(testHelper.Ctx)
		require.NoError(t, err)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser_Returns_UnauthorizedError_When_User_Is_Not_Authenticated(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update
		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		// Login with the user to get the tokens (not used in request)
		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "no authorization header provided")

		return nil
	})
}

func TestHandleUpdateUser_Returns_UnauthorizedError_When_Token_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		// Create a user to update (so the DB isn't empty)
		_, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token-here")

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "error during validation")

		return nil
	})
}

func TestHandleUpdateUser_Returns_UnauthorizedError_When_Token_Is_Expired(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		expiredToken, err := auth.MakeJWT(newUser.ID, cfg.JWTSecret, -1*time.Hour)
		require.NoError(t, err)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+expiredToken)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "error during validation")

		return nil
	})
}

func TestHandleUpdateUser_Returns_UnauthorizedError_When_Authorization_Header_Is_Malformed(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		validToken, err := auth.MakeJWT(newUser.ID, cfg.JWTSecret, time.Hour)
		require.NoError(t, err)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", validToken) // Missing "Bearer " prefix

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "invalid authorization header format")

		return nil
	})
}

func TestHandleUpdateUser_Succeeds_When_Updating_To_Same_Email(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    testutils.TEST_EMAIL,
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.UpdateUserResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)
		require.Equal(t, testutils.TEST_EMAIL, responseBody.Email)
		require.Equal(t, false, responseBody.IsChirpyRed)

		return nil
	})
}

func TestHandleUpdateUser_Returns_BadRequest_When_Email_Is_Empty(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "",
			Password: "new password",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleUpdateUser_Returns_BadRequest_When_Password_Is_Empty(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		newUser, err := app.RouteHandlers.Services.UserService.CreateUser(testHelper.Ctx, services.CreateUserInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)

		loginOutput, err := app.RouteHandlers.Services.AuthService.Login(testHelper.Ctx, services.LoginInput{
			Email:    testutils.TEST_EMAIL,
			Password: testutils.TEST_PASSWORD,
		})
		require.NoError(t, err)
		require.Equal(t, newUser.ID, loginOutput.UserID)
		require.Equal(t, newUser.Email, loginOutput.Email)
		require.Equal(t, newUser.IsChirpyRed, loginOutput.IsChirpyRed)

		body, err := json.Marshal(models.UpdateUserResource{
			Email:    "new@email.co.uk",
			Password: "",
		})
		require.NoError(t, err)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("PUT", "/api/users", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+loginOutput.Token)

		app.Router.ServeHTTP(recorder, req)
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}
