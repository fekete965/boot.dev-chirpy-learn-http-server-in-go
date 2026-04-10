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
			name   string
			method string
			url    string
		}

		fakeUUID := uuid.New().String()

		// This is a "missing methods" list, requests that should result in 405
		testCases := []tc{
			// GET /admin/metrics - covered
			{name: "POST /admin/metrics", method: http.MethodPost, url: "/admin/metrics"},
			{name: "PUT /admin/metrics", method: http.MethodPut, url: "/admin/metrics"},
			{name: "PATCH /admin/metrics", method: http.MethodPatch, url: "/admin/metrics"},
			{name: "DELETE /admin/metrics", method: http.MethodDelete, url: "/admin/metrics"},

			// POST /admin/reset - covered
			{name: "GET /admin/reset", method: http.MethodGet, url: "/admin/reset"},
			{name: "PUT /admin/reset", method: http.MethodPut, url: "/admin/reset"},
			{name: "PACH /admin/reset", method: http.MethodPatch, url: "/admin/reset"},
			{name: "DELETE /admin/reset", method: http.MethodDelete, url: "/admin/reset"},

			// GET /api/healthz - covered
			{name: "POST /api/healthz", method: http.MethodPost, url: "/api/healthz"},
			{name: "PUT /api/healthz", method: http.MethodPut, url: "/api/healthz"},
			{name: "PATCH /api/healthz", method: http.MethodPatch, url: "/api/healthz"},
			{name: "DELETE /api/healthz", method: http.MethodDelete, url: "/api/healthz"},

			// GET+POST /api/chirps - covered
			{name: "PUT /api/chirps", method: http.MethodPut, url: "/api/chirps"},
			{name: "PATCH /api/chirps", method: http.MethodPatch, url: "/api/chirps"},
			{name: "DELETE /api/chirps", method: http.MethodDelete, url: "/api/chirps"},

			// GET+DELETE /api/chirps/{chirpID} - covered
			{name: "POST /api/chirps/<CHIRP_ID>", method: http.MethodPost, url: "/api/chirps/" + fakeUUID},
			{name: "PUT /api/chirps/<CHIRP_ID>", method: http.MethodPut, url: "/api/chirps/" + fakeUUID},
			{name: "PATCH /api/chirps/<CHIRP_ID>", method: http.MethodPatch, url: "/api/chirps/" + fakeUUID},

			// POST+PUT /api/users - covered
			{name: "GET /api/users", method: http.MethodGet, url: "/api/users"},
			{name: "PATCH /api/users", method: http.MethodPatch, url: "/api/users"},
			{name: "DELETE /api/users", method: http.MethodDelete, url: "/api/users"},

			// POST /api/login - covered
			{name: "GET /api/login", method: http.MethodGet, url: "/api/login"},
			{name: "PUT /api/login", method: http.MethodPut, url: "/api/login"},
			{name: "PATCH /api/login", method: http.MethodPatch, url: "/api/login"},
			{name: "DELETE /api/login", method: http.MethodDelete, url: "/api/login"},

			// POST /api/refresh - covered
			{name: "GET /api/refresh", method: http.MethodGet, url: "/api/refresh"},
			{name: "PUT /api/refresh", method: http.MethodPut, url: "/api/refresh"},
			{name: "PATCH /api/refresh", method: http.MethodPatch, url: "/api/refresh"},
			{name: "DELETE /api/refresh", method: http.MethodDelete, url: "/api/refresh"},

			// POST /api/revoke - covered
			{name: "GET /api/revoke", method: http.MethodGet, url: "/api/revoke"},
			{name: "PUT /api/revoke", method: http.MethodPut, url: "/api/revoke"},
			{name: "PATCH /api/revoke", method: http.MethodPatch, url: "/api/revoke"},
			{name: "DELETE /api/revoke", method: http.MethodDelete, url: "/api/revoke"},

			// POST /api/polka/webhooks - covered
			{name: "GET /api/polka/webhooks", method: http.MethodGet, url: "/api/polka/webhooks"},
			{name: "PUT /api/polka/webhooks", method: http.MethodPut, url: "/api/polka/webhooks"},
			{name: "PATCH /api/polka/webhooks", method: http.MethodPatch, url: "/api/polka/webhooks"},
			{name: "DELETE /api/polka/webhooks", method: http.MethodDelete, url: "/api/polka/webhooks"},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				recorder := httptest.NewRecorder()
				req := httptest.NewRequest(tc.method, tc.url, nil)

				app.Router.ServeHTTP(recorder, req)

				require.Equal(t, http.StatusMethodNotAllowed, recorder.Code, "%s %s", tc.method, tc.url)
				require.Equal(t, "text/plain; charset=utf-8", recorder.Header().Get("Content-Type"), "%s %s", tc.method, tc.url)
				require.Equal(t, "Method Not Allowed\n", recorder.Body.String(), "%s %s", tc.method, tc.url)
			})
		}

		return nil
	})
}
