package router

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	handlersPkg "github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/handlers"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testdb"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func setupAdminRouteTestData(t *testing.T, q *database.Queries, ctx context.Context) *database.User {
	generator := testdb.NewGenerator(testdb.NewGeneratorInput{
		Db: q,
		T:  t,
	})
	user := generator.GenerateUser(ctx)
	generator.GenerateChirps(ctx, user.ID, 5)

	return user
}

func TestHandleReset_InProductionMode(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)

	cfg := testutils.GetTestApiConfig()
	cfg.Platform = "production"

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		generator := testdb.NewGenerator(testdb.NewGeneratorInput{
			Db: q,
			T:  t,
		})
		user := generator.GenerateUser(testHelper.Ctx)
		generator.GenerateChirps(testHelper.Ctx, user.ID, 5)

		// Setup metrics
		cfg.FileserverHits.Add(10)

		app := newRouterTestApp(cfg, q)

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

		app.Router.ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusForbidden, recorder.Code)
		require.Equal(t, "Forbidden operation", recorder.Body.String())

		// Check database and metrics unchanged
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
	testHelper := testutils.NewTestHelper(t)

	cfg := testutils.GetTestApiConfig()
	cfg.Platform = "dev"

	testHelper.WithTx(func(q *database.Queries) error {
		// Setup test data
		setupAdminRouteTestData(t, q, testHelper.Ctx)

		// Setup metrics
		cfg.FileserverHits.Add(10)

		app := newRouterTestApp(cfg, q)

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

		app.Router.ServeHTTP(recorder, req)

		// Check response
		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "Metric has been reset", recorder.Body.String())

		// Check database and metrics reset
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
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		user := setupAdminRouteTestData(t, q, testHelper.Ctx)

		validPayload, err := json.Marshal(models.WebhookResource{
			Event: "user.upgraded",
			Data: models.WebhookEventData{
				UserID: user.ID,
			},
		})
		require.NoError(t, err)

		initialUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, false, initialUserData.IsChirpyRed)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey "+cfg.PolkaWebhookSecret)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNoContent, recorder.Code)

		updatedUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, true, updatedUserData.IsChirpyRed)

		return nil
	})
}

func TestHandlePolkaWebhooks_WithUnhandledEvent(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		user := setupAdminRouteTestData(t, q, testHelper.Ctx)

		validPayload, err := json.Marshal(models.WebhookResource{
			Event: "unhandled-event-type",
			Data: models.WebhookEventData{
				UserID: user.ID,
			},
		})
		require.NoError(t, err)

		initialUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, false, initialUserData.IsChirpyRed)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey "+cfg.PolkaWebhookSecret)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNoContent, recorder.Code)

		updatedUserData, err := q.FindUserById(testHelper.Ctx, user.ID)
		require.NoError(t, err)
		require.Equal(t, false, updatedUserData.IsChirpyRed)

		return nil
	})
}

func TestHandlePolkaWebhooks_WithMissingUserID(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		validPayload, err := json.Marshal(models.WebhookResource{
			Event: "user.upgraded",
			Data: models.WebhookEventData{
				UserID: uuid.New(),
			},
		})
		require.NoError(t, err)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", bytes.NewBuffer(validPayload))
		req.Header.Set("Authorization", "ApiKey "+cfg.PolkaWebhookSecret)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusNotFound, recorder.Code)
		require.Equal(t, "user not found", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithoutAuthorizationHeader(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		setupAdminRouteTestData(t, q, testHelper.Ctx)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "no authorization header provided", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidAuthorizationHeaderFormat(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		setupAdminRouteTestData(t, q, testHelper.Ctx)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)
		req.Header.Set("Authorization", "FromApiKeyFormat this-is-a-token")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "invalid authorization header format", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidApiKey(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		setupAdminRouteTestData(t, q, testHelper.Ctx)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", nil)
		req.Header.Set("Authorization", "ApiKey this-is-an-invalid-api-key")

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusUnauthorized, recorder.Code)
		require.Equal(t, "invalid api key", recorder.Body.String())

		return nil
	})
}

func TestHandlePolkaWebhooks_WithInvalidPayload(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		setupAdminRouteTestData(t, q, testHelper.Ctx)

		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("POST", "/api/polka/webhooks", strings.NewReader("invalid payload"))
		req.Header.Set("Authorization", "ApiKey "+cfg.PolkaWebhookSecret)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusBadRequest, recorder.Code)
		require.Equal(t, "error decoding request body", recorder.Body.String())

		return nil
	})
}

// This import is kept to ensure we compile against the handlers package;
// it also avoids accidental name shadowing when adding new tests.
var _ = handlersPkg.Handlers{}
