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
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/router"
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
		Cfg: cfg,
		Db: database.New(dbConnection),
	})

	routeHandlers := handlers.NewHandlers(handlers.NewHandlersInput{
		Cfg: cfg,
		Services: newServices,
	})

	muxRouter := router.GetNewRouter(router.GetNewRouterInput{
		Cfg: cfg,
		RouteHandlers: routeHandlers,
	})

	addr := fmt.Sprintf(":%d", cfg.Port)
	server := http.Server{
		Addr: addr,
		Handler: muxRouter,
	}

	err = server.ListenAndServe()
	if err != nil {
		log.Fatalf("could not start the server: %v", err)
	}

	fmt.Printf("server started on port: %d\n", envVars.Port)
}
