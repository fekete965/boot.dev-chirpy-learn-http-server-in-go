package services

import (
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
)

type Services struct {
	UserService  UserService
	ChirpService ChirpService
	AuthService  AuthService
}

type NewServicesInput struct {
	Cfg *config.ApiConfig
	Db  *database.Queries
}

func NewServices(input NewServicesInput) *Services {
	userService := NewUserService(input.Db)
	chirpService := NewChirpService(NewChirpServiceInput{
		Cfg:         input.Cfg,
		Db:          input.Db,
		UserService: userService,
	})
	authService := NewAuthService(NewAuthServiceInput{
		Cfg:         input.Cfg,
		Db:          input.Db,
		UserService: userService,
	})

	return &Services{
		UserService:  userService,
		ChirpService: chirpService,
		AuthService:  authService,
	}
}
