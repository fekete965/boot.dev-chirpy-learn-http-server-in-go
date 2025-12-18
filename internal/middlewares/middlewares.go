package middlewares

import "github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"

type Middlewares struct {
	cfg *config.ApiConfig
}

type NewMiddlewaresInput struct {
	cfg *config.ApiConfig
}

func NewMiddlewares(input NewMiddlewaresInput) *Middlewares {
	return &Middlewares{
		cfg: input.cfg,
	}
}
