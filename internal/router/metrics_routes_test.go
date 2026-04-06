package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/stretchr/testify/require"
)

func TestAdminMetrics_Returns_HTML_With_Current_Count(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	// Pretend we've served the static site 3 times
	cfg.FileserverHits.Add(3)

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		recorder := httptest.NewRecorder()
		req := httptest.NewRequest("GET", "/admin/metrics", nil)

		app.Router.ServeHTTP(recorder, req)

		require.Equal(t, http.StatusOK, recorder.Code)
		require.Equal(t, "text/html; charset=utf-8", recorder.Header().Get("Content-Type"))
		require.Contains(t, recorder.Body.String(), "Chirpy has been visited 3 times!")

		return nil
	})
}
