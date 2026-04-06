package handlers

import (
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
)

type Handlers struct {
	Cfg      *config.ApiConfig
	Services *services.Services
}

type NewHandlersInput struct {
	Cfg      *config.ApiConfig
	Services *services.Services
}

func NewHandlers(input NewHandlersInput) *Handlers {
	return &Handlers{
		Cfg:      input.Cfg,
		Services: input.Services,
	}
}

func GetHandlers(cfg *config.ApiConfig, q *database.Queries) *Handlers {
	newServices := services.NewServices(services.NewServicesInput{
		Cfg: cfg,
		Db:  q,
	})

	handlers := NewHandlers(NewHandlersInput{
		Cfg:      cfg,
		Services: newServices,
	})

	return handlers
}
