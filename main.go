package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"sync/atomic"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/handlers"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
)


func main() {
	envVars, err := config.LoadEnv()
	if err != nil {
		log.Fatalf("error loading environment variables: %v", err)
	}

	dbConnection, err := sql.Open("postgres", envVars.DbUrl)
	if err != nil {
		log.Fatalf("error connecting to the database: %v", err)
	}

	cfg := &config.ApiConfig{
		FileserverHits: atomic.Int32{},
		JWTSecret: envVars.JWTSecret,
		Platform: envVars.Platform,
		PolkaWebhookSecret: envVars.PolkaWebhookSecret,
		Port: envVars.Port,
	}

	newServices := services.NewServices(services.NewServicesInput{
		cfg: cfg,
		db: database.New(dbConnection),
	})

	routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
		cfg: cfg,
		services: newServices,
	})

	apiMiddlewares := middlewares.NewMiddlewares(middlewares.NewMiddlewaresInput{
		cfg: cfg,
	})

	mux := http.NewServeMux()

	mux.Handle("GET /admin/metrics", apiMiddlewares.MiddlewareHandleMetrics())
	mux.Handle("POST /admin/reset", routeHandlers.HandleReset())
	mux.Handle("GET /api/healthz", handlers.HandleHealthCheck)
	mux.Handle("POST /api/chirps", routeHandlers.HandleCreateChirp())
	mux.Handle("GET /api/chirps", routeHandlers.HandleGetAllChirps())
	mux.Handle("GET /api/chirps/{chirpID}", routeHandlers.HandleGetChirpById())
	mux.Handle("DELETE /api/chirps/{chirpID}", routeHandlers.HandleDeleteChirp())
	mux.Handle("POST /api/users", routeHandlers.HandleCreateUser())
	mux.Handle("PUT /api/users", routeHandlers.HandleUpdateUser())
	mux.Handle("POST /api/login", routeHandlers.HandleLogin())
	mux.Handle("POST /api/refresh", routeHandlers.HandleTokenRefresh())
	mux.Handle("POST /api/revoke", routeHandlers.HandleTokenRevoke())
	mux.Handle("POST /api/polka/webhooks", routeHandlers.HandlePolkaWebhooks())
	mux.Handle("/app/assets/", apiMiddlewares.MiddlewareMetricsInc(handlers.HandleAssets))
	mux.Handle("/app/", apiMiddlewares.MiddlewareMetricsInc(handlers.HandleHome))

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := http.Server{
		Addr: addr,
		Handler: mux,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("could not start the server: %v", err)
	}

	fmt.Printf("server started on port: %d\n", envVars.Port)
}
