package handlers

import (
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
)

type Handlers struct {
	cfg *config.ApiConfig
	services *services.Services
}

type NewHandlersInput struct {
	cfg *config.ApiConfig
	services *services.Services
}

func NewHandlers(input NewHandlersInput) *Handlers {
	return &Handlers{
		cfg: input.cfg,
		services: input.services,
	}
}
