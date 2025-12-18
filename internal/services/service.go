package services

import (
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
)


type Services struct {
	UserService UserService
	ChirpService ChirpService
	AuthService AuthService
}

type NewServicesInput struct {
	cfg *config.ApiConfig
	db *database.Queries
}

func NewServices(input NewServicesInput) *Services {
	userService := NewUserService(input.db)
	chirpService := NewChirpService(NewChirpServiceInput{
		cfg: input.cfg,
		db: input.db,
		userService: userService,
	})
	authService := NewAuthService(NewAuthServiceInput{
		cfg: input.cfg,
		db: input.db,
		userService: userService,
	})

	return &Services{
		UserService: userService,
		ChirpService: chirpService,
		AuthService: authService,
	}
}
