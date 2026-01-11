package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testdb"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func setupTestData(t *testing.T, q *database.Queries, ctx context.Context) *database.User {
	generator := testdb.NewGenerator(testdb.NewGeneratorInput{
		Db: q,
		T: t,
	})
	user := generator.GenerateUser(ctx)
	generator.GenerateChirps(ctx, user.ID, 5)

	return user
}

func TestHandleReset_InProductionMode(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()
	cfg.Platform = "production"

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T: t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		// Setup metrics
		cfg.FileserverHits.Add(10)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup base cases
		initialUserCount, err := q.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), initialUserCount)
		
		initialChirpCount, err := q.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(5), initialChirpCount)

		initialFileServerHits := cfg.FileserverHits.Load()
		require.Equal(t, int32(10), initialFileServerHits)
		
		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/reset", nil)
		require.NoError(t, err)

		// Handle the request
		handlers.HandleReset().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusForbidden, recorder.Code)
		require.Equal(t, "Forbidden operation", recorder.Body.String())

		// Check database and metrics
		updatedUserCount, err := q.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, updatedUserCount, initialUserCount)
		
		updatedChirpCount, err := q.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, updatedChirpCount, initialChirpCount)

		updatedFileServerHits := cfg.FileserverHits.Load()
		require.Equal(t, updatedFileServerHits, initialFileServerHits)

		return nil
	})
}

func TestHandleReset_InDevelopmentMode(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()
	cfg.Platform = "dev"

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupTestData(t, q, testHelper.Ctx)

		// Setup metrics
		cfg.FileserverHits.Add(10)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup base cases
		initialUserCount, err := q.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(1), initialUserCount)
		
		initialChirpCount, err := q.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(5), initialChirpCount)

		initialFileServerHits := cfg.FileserverHits.Load()
		require.Equal(t, int32(10), initialFileServerHits)
		
		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/admin/reset", nil)
		require.NoError(t, err)

		// Handle the request
		handlers.HandleReset().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "Metric has been reset", recorder.Body.String())

		// Check database and metrics
		updatedUserCount, err := q.GetUserCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), updatedUserCount)
		
		updatedChirpCount, err := q.GetChirpCount(testHelper.Ctx)
		require.NoError(t, err)
		require.Equal(t, int64(0), updatedChirpCount)

		updatesFileServerHits := cfg.FileserverHits.Load()
		require.Equal(t, int32(0), updatesFileServerHits)

		return nil
	})
}

func TestHandlePolkaWebhooks(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		user := setupTestData(t, q, testHelper.Ctx)

		validPayload, err := json.Marshal(models.WebhookResource {
			Event: "user.upgraded",
			Data: models.WebhookEventData{
				UserID: user.ID,
			},
		})
		require.NoError(t, err)

		// Setup the base case
		initialUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, initialUserData.IsChirpyRed, false)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey " + cfg.PolkaWebhookSecret)
		require.NoError(t, err)

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusNoContent, recorder.Code)
		
		// Check updated user data
		updatedUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, updatedUserData.IsChirpyRed, true)

		return nil
	})
}

func TestHandlePolkaWebhooks_WithUnhandledEvent(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		user := setupTestData(t, q, testHelper.Ctx)

		validPayload, err := json.Marshal(models.WebhookResource {
			Event: "unhandled-event-type",
			Data: models.WebhookEventData{
				UserID: user.ID,
			},
		})
		require.NoError(t, err)

		// Setup the base case
		initialUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, initialUserData.IsChirpyRed, false)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey " + cfg.PolkaWebhookSecret)
		require.NoError(t, err)

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusNoContent, recorder.Code)
		
		// Check updated user data
		updatedUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, updatedUserData.IsChirpyRed, false)

		return nil
	})
}

func TestHandlePolkaWebhooks_WithMissingUserID(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {

		validPayload, err := json.Marshal(models.WebhookResource {
			Event: "user.upgraded",
			Data: models.WebhookEventData{
				UserID: uuid.New(),
			},
		})
		require.NoError(t, err)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey " + cfg.PolkaWebhookSecret)

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithoutAuthorizationHeader(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupTestData(t, q, testHelper.Ctx)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)

		// We don't set the API key in the request header
		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "no authorization header provided", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidAuthorizationHeaderFormat(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupTestData(t, q, testHelper.Ctx)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)

		// We set the API key in the request header but using the wrong format
		req.Header.Set("Authorization", "FromApiKeyFormat this-is-a-token")

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "invalid authorization header format", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidApiKey(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupTestData(t, q, testHelper.Ctx)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)

		// We set the API key in the request header with an invalid API key
		req.Header.Set("Authorization", "ApiKey this-is-an-invalid-api-key")

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "invalid api key", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidPayload(t *testing.T) {
	testHelper := testutils.SetupServiceTest(t)

	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupTestData(t, q, testHelper.Ctx)

		// Setup handlers
		handlers := GetHandlers(cfg, q)

		// Setup request
		recorder := httptest.NewRecorder()
		// Set an invalid payload format
		req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader("invalid payload"))
		req.Header.Set("Authorization", "ApiKey " + cfg.PolkaWebhookSecret)

		handlers.HandlePolkaWebhooks().ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}
