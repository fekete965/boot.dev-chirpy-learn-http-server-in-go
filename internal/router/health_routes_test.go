package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestHandleHealthCheck(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/api/healthz", nil)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "OK", recorder.Body.String())

		return nil
	})
}
