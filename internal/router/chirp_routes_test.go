package router

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/handlers"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testdb"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func createAndLoginUser(t *testing.T, ctx context.Context, handlers *handlers.Handlers) (*services.User, *services.LoginOutput) {
	// Create a new user
	user, err := handlers.Services.UserService.CreateUser(ctx, services.CreateUserInput{
		Email: testutils.TEST_EMAIL,
		Password: testutils.TEST_PASSWORD,
	})
	require.NoError(t, err)

	// Login with the user
	loginOutput, err := handlers.Services.AuthService.Login(ctx, services.LoginInput{
		Email: user.Email,
		Password: testutils.TEST_PASSWORD,
	})
	require.NoError(t, err)

	return &user, &loginOutput
}

func TestHandleCreateChirp(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})
		
		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: "test chirp",
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusCreated, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.CreateChirpResponse
		err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, user.ID, responseBody.UserID)
		require.Equal(t, "test chirp", responseBody.Body)
		require.False(t, responseBody.CreatedAt.IsZero())
		require.False(t, responseBody.UpdatedAt.IsZero())

		return nil
	})
}

func TestHandleCreateChirp_Profanity_Filtering(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})
		
		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)


		type profanityTestCase struct {
			chirpBody string
			expectedChirpBody string
		}

		profanityTestCases := make([]profanityTestCase, len(constants.PROFANE_WORDS))
		for i, word := range constants.PROFANE_WORDS {
			profanityTestCases[i] = profanityTestCase{
				chirpBody: fmt.Sprintf("This is a %s great day!", word),
				expectedChirpBody: "This is a **** great day!",
			}
			profanityTestCases[i] = profanityTestCase{
				chirpBody: fmt.Sprintf("This is a %s!", word),
				expectedChirpBody: fmt.Sprintf("This is a %s!", word),
			}
		}

		for _, testCase := range profanityTestCases {
			payload, err := json.Marshal(models.CreateChirpResource{
				Body: testCase.chirpBody,
			})
			require.NoError(t, err)

			// Setup the request
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
			req.Header.Set("Content-Type", "application/json")
			req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

			// Handle request
			router.ServeHTTP(recorder, req)

			require.Equal(t, http.StatusCreated, recorder.Code)
			require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

			var responseBody models.CreateChirpResponse
			err = json.Unmarshal(recorder.Body.Bytes(), &responseBody)
			require.NoError(t, err)

			require.Equal(t, user.ID, responseBody.UserID)
			require.Equal(t, testCase.expectedChirpBody, responseBody.Body)
			require.False(t, responseBody.CreatedAt.IsZero())
			require.False(t, responseBody.UpdatedAt.IsZero())
		}

		return nil
	})
}

func TestHandleCreateChirp_Returns_BadRequestError_When_Payload_Is_Malformed(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		_, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", strings.NewReader("invalid payload"))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateChirp_Returns_BadRequestError_When_Body_Is_Empty(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		_, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: "",
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "validation error", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateChirp_Returns_BadRequestError_When_Body_Is_Too_Long(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		_, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: strings.Repeat("a", constants.MAX_CHIRP_LENGTH + 1),
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "invalid chirp length", recorder.Body.String())

		return nil
	})
}

func TestHandleCreateChirp_Returns_UnauthorizedError_Without_Authorization_Header(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: "test chirp",
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "no authorization header provided")

		return nil
	})
}

func TestHandleCreateChirp_Returns_UnauthorizedError_With_Invalid_Token(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: "test chirp",
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer invalid-token")

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "error validating token")

		return nil
	})
}

func TestHandleCreateChirp_Returns_NotFoundError_When_User_Is_Not_Found(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		_, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)

		payload, err := json.Marshal(models.CreateChirpResource{
			Body: "Hello, there!",
		})
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/chirps", bytes.NewReader(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Delete the user
		err = newServices.UserService.DeleteAllUsers(testHelper.Ctx)
		require.NoError(t, err)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandleGetAllChirps(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 10, len(responseBody))

		// Check that the chirps are sorted by created_at in ascending order (default sort)
		for i := 0; i < len(responseBody) - 1; i++ {
			currentChirp := responseBody[i]
			nextChirp := responseBody[i + 1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestHandleGetAllChirps_With_No_Chirps_Present_In_The_Database(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 0, len(responseBody))

		return nil
	})
}

func TestHandleGetAllChirps_For_A_Specific_User(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add the author_id query parameter
		query := req.URL.Query()
		query.Add("author_id", user1.ID.String())
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 5, len(responseBody))

		for _, chirp := range responseBody {
			require.Equal(t, user1.ID, chirp.UserID)
		}

		return nil
	})
}

func TestHandleGetAllChirps_For_A_Specific_User_Who_Has_No_Chirps(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add the author_id query parameter
		query := req.URL.Query()
		query.Add("author_id", user2.ID.String())
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 0, len(responseBody))

		return nil
	})
}

func TestHandleGetAllChirps_With_Sort_Created_At_Asc(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add the sort query parameter to sort by created_at in ascending order
		query := req.URL.Query()
		query.Add("sort", "asc")
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 10, len(responseBody))

		for i := 0; i < len(responseBody) - 1; i++ {
			currentChirp := responseBody[i]
			nextChirp := responseBody[i + 1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestHandleGetAllChirps_With_Sort_Created_At_Desc(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add the sort query parameter to sort by created_at in descending order
		query := req.URL.Query()
		query.Add("sort", "desc")
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 10, len(responseBody))

		for i := 0; i < len(responseBody) - 1; i++ {
			currentChirp := responseBody[i]
			nextChirp := responseBody[i + 1]

			require.True(t, currentChirp.CreatedAt.After(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestHandleGetAllChirps_With_Invalid_Sort_Parameter_Defaults_To_Asc(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1 := generator.GenerateUser(testHelper.Ctx)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add an invalid sort query parameter, we should default to ascending order
		query := req.URL.Query()
		query.Add("sort", "invalid-sort-parameter")
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody []models.GetAllChirpResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, 10, len(responseBody))

		for i := 0; i < len(responseBody) - 1; i++ {
			currentChirp := responseBody[i]
			nextChirp := responseBody[i + 1]

			require.True(t, currentChirp.CreatedAt.Before(nextChirp.CreatedAt))
		}

		return nil
	})
}

func TestHandleGetAllChirps_Returns_BadRequestError_When_Author_ID_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps", nil)
		
		// Add the author_id query parameter with an unknown user id
		query := req.URL.Query()
		query.Add("author_id", "invalid-author-id")
		req.URL.RawQuery = query.Encode()

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "invalid author_id")

		return nil
	})
}

func TestHandleGetChirpById(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		targetChirp := chirps[0]

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps/" + targetChirp.ID.String(), nil)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "application/json", recorder.Header().Get("Content-Type"))

		var responseBody models.GetChirpByIdResponse
		err := json.Unmarshal(recorder.Body.Bytes(), &responseBody)
		require.NoError(t, err)

		require.Equal(t, targetChirp.ID, responseBody.ID)
		require.Equal(t, user.ID, responseBody.UserID)
		require.Equal(t, targetChirp.Body, responseBody.Body)
		require.True(t, targetChirp.CreatedAt.Equal(responseBody.CreatedAt))
		require.True(t, targetChirp.UpdatedAt.Equal(responseBody.UpdatedAt))

		return nil
	})
}

func TestHandleGetChirpById_Returns_BadRequestError_When_Chirp_ID_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps/" + "invalid-chirp-id", nil)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "cannot parse chirpID")

		return nil
	})
}

func TestHandleGetChirpById_Returns_NotFoundError_When_Chirp_Is_Not_Found(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
			newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/chirps/" + uuid.New().String(), nil)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "chirp not found", recorder.Body.String())

		return nil
	})
}

func TestHandleDeleteChirp(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		targetChirp := chirps[0]

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + targetChirp.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNoContent, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))

		// Check that the chirp was deleted
		_, err := q.GetChirpById(testHelper.Ctx, targetChirp.ID)
		require.Error(t, err)
		require.True(t, errors.Is(err, sql.ErrNoRows))

		return nil
	})
}

func TestHandleDeleteChirp_Returns_BadRequestError_When_Chirp_ID_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/invalid-chirp-id", nil)
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "cannot parse chirpID")

		return nil
	})
}

func TestHandleDeleteChirp_Returns_NotFoundError_When_Chirp_ID_Not_Found(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + uuid.New().String(), nil)
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "chirp not found", recorder.Body.String())

		return nil
	})
}

func TestHandleDeleteChirp_Returns_NotFoundError_When_User_Is_Missing(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user, loginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)
		targetChirp := chirps[0]

		// Delete the user
		err := newServices.UserService.DeleteAllUsers(testHelper.Ctx)
		require.NoError(t, err)

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + targetChirp.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer " + loginOutput.Token)

		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandleDeleteChirp_Returns_UnauthorizedError_When_User_Is_Not_Authenticated(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)
		targetChirp := chirps[0]

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + targetChirp.ID.String(), nil)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "no authorization header provided")

		return nil
	})
}

func TestHandleDeleteChirp_Returns_UnauthorizedError_When_User_Is_Unauthorized(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user1, user1LoginOutput := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		user2 := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user1.ID, 5)
		user2Chirps :=generator.GenerateChirps(testHelper.Ctx, user2.ID, 5)

		targetChirp := user2Chirps[0]

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + targetChirp.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer " + user1LoginOutput.Token)
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusForbidden, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Equal(t, "invalid user permission", recorder.Body.String())

		return nil
	})
}

func TestHandleDeleteChirp_Returns_UnauthorizedError_When_SessionToken_Is_Invalid(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user, _ := createAndLoginUser(t, testHelper.Ctx, routeHandlers)
		chirps := generator.GenerateChirps(testHelper.Ctx, user.ID, 5)
		targetChirp := chirps[0]

		// Setup the request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("DELETE", "/api/chirps/" + targetChirp.ID.String(), nil)
		req.Header.Set("Authorization", "Bearer invalid-token")
		
		// Handle request
		router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "error validating token")

		return nil
	})
}

func TestUnsupportedRequestMethods(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		newServices := services.NewServices(services.NewServicesInput{
			Cfg: cfg,
			Db: q,
		})

		routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
			Cfg: cfg,
			Services: newServices,
		})

		router := GetNewRouter(GetNewRouterInput{
			RouteHandlers: routeHandlers,
			Cfg: cfg,
		})

		type unsupportedRequestMethodTestCase struct {
			method string
			url string
		}

		chirpID := uuid.New().String()
		testCases := []unsupportedRequestMethodTestCase{
			// /api/chirps
			{method: http.MethodPut, url: "/api/chirps"},
			{method: http.MethodPatch, url: "/api/chirps"},
			{method: http.MethodDelete, url: "/api/chirps"},

			// /api/chirps/{chirpID} 
			{method: http.MethodPost, url: "/api/chirps/" + chirpID},
			{method: http.MethodPut, url: "/api/chirps/" + chirpID},
			{method: http.MethodPatch, url: "/api/chirps/" + chirpID},
		}

		for _, tc := range testCases {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.url, nil)

			router.ServeHTTP(recorder, req)

			require.Equal(t, http.StatusMethodNotAllowed, recorder.Code)
			require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"))
			require.Equal(t, "Method Not Allowed\n", recorder.Body.String())
		}

		return nil
	})
}
