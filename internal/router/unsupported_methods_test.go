package router

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/require"
)

func TestUnsupportedRequestMethods_Router(t *testing.T) {
	testHelper := testutils.NewTestHelper(t)
	cfg := testutils.GetTestApiConfig()

	testHelper.WithTx(func(q *database.Queries) error {
		app := newRouterTestApp(cfg, q)

		type tc struct {
			method string
			url    string
		}

		fakeUUID := uuid.New().String()

		// This is a "missing methods" list, requests that should result in 405
		testCases := []tc{
			// GET /admin/metrics - covered
			{method: http.MethodPost, url: "/admin/metrics"},
			{method: http.MethodPut, url: "/admin/metrics"},
			{method: http.MethodPatch, url: "/admin/metrics"},
			{method: http.MethodDelete, url: "/admin/metrics"},

			// POST /admin/reset - covered
			{method: http.MethodGet, url: "/admin/reset"},
			{method: http.MethodPut, url: "/admin/reset"},
			{method: http.MethodPatch, url: "/admin/reset"},
			{method: http.MethodDelete, url: "/admin/reset"},

			// GET /api/healthz - covered
			{method: http.MethodPost, url: "/api/healthz"},
			{method: http.MethodPut, url: "/api/healthz"},
			{method: http.MethodPatch, url: "/api/healthz"},
			{method: http.MethodDelete, url: "/api/healthz"},

			// GET+POST /api/chirps - covered
			{method: http.MethodPut, url: "/api/chirps"},
			{method: http.MethodPatch, url: "/api/chirps"},
			{method: http.MethodDelete, url: "/api/chirps"},

			// GET+DELETE /api/chirps/{chirpID} - covered
			{method: http.MethodPost, url: "/api/chirps/" + fakeUUID},
			{method: http.MethodPut, url: "/api/chirps/" + fakeUUID},
			{method: http.MethodPatch, url: "/api/chirps/" + fakeUUID},

			// POST+PUT /api/users - covered
			{method: http.MethodGet, url: "/api/users"},
			{method: http.MethodPatch, url: "/api/users"},
			{method: http.MethodDelete, url: "/api/users"},

			// POST /api/login - covered
			{method: http.MethodGet, url: "/api/login"},
			{method: http.MethodPut, url: "/api/login"},
			{method: http.MethodPatch, url: "/api/login"},
			{method: http.MethodDelete, url: "/api/login"},

			// POST /api/refresh - covered
			{method: http.MethodGet, url: "/api/refresh"},
			{method: http.MethodPut, url: "/api/refresh"},
			{method: http.MethodPatch, url: "/api/refresh"},
			{method: http.MethodDelete, url: "/api/refresh"},

			// POST /api/revoke - covered
			{method: http.MethodGet, url: "/api/revoke"},
			{method: http.MethodPut, url: "/api/revoke"},
			{method: http.MethodPatch, url: "/api/revoke"},
			{method: http.MethodDelete, url: "/api/revoke"},

			// POST /api/polka/webhooks - covered
			{method: http.MethodGet, url: "/api/polka/webhooks"},
			{method: http.MethodPut, url: "/api/polka/webhooks"},
			{method: http.MethodPatch, url: "/api/polka/webhooks"},
			{method: http.MethodDelete, url: "/api/polka/webhooks"},
		}

		for _, tc := range testCases {
			recorder := httptest.NewRecorder()
			req := httptest.NewRequest(tc.method, tc.url, nil)

			app.Router.ServeHTTP(recorder, req)

			require.Equal(t, http.StatusMethodNotAllowed, recorder.Code, "%s %s", tc.method, tc.url)
			require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"), "%s %s", tc.method, tc.url)
			require.Equal(t, "Method Not Allowed\n", recorder.Body.String(), "%s %s", tc.method, tc.url)
		}

		return nil
	})
}
