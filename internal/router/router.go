package router

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/handlers"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
)

type GetNewRouterInput struct {
	Cfg           *config.ApiConfig
	RouteHandlers *handlers.Handlers
}

func GetNewRouter(input GetNewRouterInput) *http.ServeMux {
	apiMiddlewares := middlewares.NewMiddlewares(middlewares.NewMiddlewaresInput{
		Cfg: input.Cfg,
	})

	mux := http.NewServeMux()

	mux.Handle("GET /admin/metrics", apiMiddlewares.MiddlewareHandleMetrics())
	mux.Handle("POST /admin/reset", input.RouteHandlers.HandleReset())
	mux.Handle("GET /api/healthz", handlers.HandleHealthCheck)
	mux.Handle("POST /api/chirps", input.RouteHandlers.HandleCreateChirp())
	mux.Handle("GET /api/chirps", input.RouteHandlers.HandleGetAllChirps())
	mux.Handle("GET /api/chirps/{chirpID}", input.RouteHandlers.HandleGetChirpById())
	mux.Handle("DELETE /api/chirps/{chirpID}", input.RouteHandlers.HandleDeleteChirp())
	mux.Handle("POST /api/users", input.RouteHandlers.HandleCreateUser())
	mux.Handle("PUT /api/users", input.RouteHandlers.HandleUpdateUser())
	mux.Handle("POST /api/login", input.RouteHandlers.HandleLogin())
	mux.Handle("POST /api/refresh", input.RouteHandlers.HandleTokenRefresh())
	mux.Handle("POST /api/revoke", input.RouteHandlers.HandleTokenRevoke())
	mux.Handle("POST /api/polka/webhooks", input.RouteHandlers.HandlePolkaWebhooks())
	mux.Handle("/app/assets/", apiMiddlewares.MiddlewareMetricsInc(handlers.HandleAssets))
	mux.Handle("/app/", apiMiddlewares.MiddlewareMetricsInc(handlers.HandleHome))

	return mux
}
