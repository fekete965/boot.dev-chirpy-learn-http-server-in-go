package middlewares

import "github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"

type Middlewares struct {
	Cfg *config.ApiConfig
}

type NewMiddlewaresInput struct {
	Cfg *config.ApiConfig
}

func NewMiddlewares(input NewMiddlewaresInput) *Middlewares {
	return &Middlewares{
		Cfg: input.Cfg,
	}
}
