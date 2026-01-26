package router

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/handlers"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
)

// routerTestApp mirrors the main wiring (cfg -> services -> handlers -> router)
// so router tests exercise real route matching and handler behavior.
type routerTestApp struct {
	Cfg           *config.ApiConfig
	Services      *services.Services
	RouteHandlers *handlers.Handlers
	Router        *http.ServeMux
}

func newRouterTestApp(cfg *config.ApiConfig, q *database.Queries) routerTestApp {
	newServices := services.NewServices(services.NewServicesInput{
		Cfg: cfg,
		Db:  q,
	})

	routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
		Cfg:      cfg,
		Services: newServices,
	})

	r := GetNewRouter(GetNewRouterInput{
		RouteHandlers: routeHandlers,
		Cfg:           cfg,
	})

	return routerTestApp{
		Cfg:           cfg,
		Services:      newServices,
		RouteHandlers: routeHandlers,
		Router:        r,
	}
}

