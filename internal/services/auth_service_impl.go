package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
)


type authService struct {
	Cfg *config.ApiConfig
	Db *database.Queries
	UserService UserService
}

type NewAuthServiceInput struct {
	Cfg *config.ApiConfig
	Db *database.Queries
	UserService UserService
}

func NewAuthService(input NewAuthServiceInput) *authService {
	return &authService{
		Cfg: input.Cfg,
		Db: input.Db,
		UserService: input.UserService,
	}
}

func (s *authService) Login(ctx context.Context, input LoginInput) (LoginOutput, error)  {
	user, err := s.UserService.FindUserByEmail(ctx, FindUserByEmailInput{Email: input.Email})
	if err != nil {
		errorMessage := "failed to find user"
		log.Printf("%s: %v", errorMessage, err)
		return LoginOutput{}, service_errors.NewInternalServerError(errorMessage)
	}

	match, err := auth.CheckPasswordHash(input.Password, user.HashedPassword)
		if err != nil {
			errorMessage := "error checking password"
			log.Printf("%s: %v", errorMessage, err)
			return LoginOutput{}, service_errors.NewInternalServerError(errorMessage)
		}

		if !match {
			errorMessage := "invalid credentials"
			log.Print(errorMessage)
			return LoginOutput{}, service_errors.NewUnauthorizedError(errorMessage)
		}
	
	token, err := auth.MakeJWT(user.ID, s.Cfg.JWTSecret, constants.DEFAULT_EXPIRES_IN)
	if err != nil {
		errorMessage := "error generating access token"
		log.Printf("%s: %v", errorMessage, err)
		return LoginOutput{}, service_errors.NewInternalServerError(errorMessage)
	}
	
	refreshTokenValue, err := auth.MakeRefreshToken()
	if err != nil {
		errorMessage := "error generating refresh token"
		log.Printf("%s: %v", errorMessage, err)
		return LoginOutput{}, service_errors.NewInternalServerError(errorMessage)
	}

	now := time.Now()
	expiresAt := now.Add(constants.REFRESH_TOKEN_EXPIRES_IN)
	refreshToken, err := s.Db.CreateRefreshToken(ctx, database.CreateRefreshTokenParams{
		Token: refreshTokenValue,
		UserID: user.ID,
		ExpiresAt: expiresAt,
		RevokedAt: sql.NullTime{},
		CreatedAt: now,
		UpdatedAt: now,
	})
	if err != nil {
		errorMessage := "error creating refresh token"
		log.Printf("%s: %v", errorMessage, err)
		return LoginOutput{}, service_errors.NewInternalServerError(errorMessage)
	}

	return LoginOutput{
		UserID: user.ID,
		Email: user.Email,
		IsChirpyRed: user.IsChirpyRed,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
		Token: token,
		RefreshToken: refreshToken.Token,
	}, nil
}

func (s *authService) RefreshToken(ctx context.Context, input RefreshTokenInput) (string, error) {
	refreshToken, err := s.Db.FindRefreshToken(ctx, input.RefreshToken)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorMessage := "refresh token not found"
			log.Print(errorMessage)
			return "", service_errors.NewNotFoundError(errorMessage)
		}

		errorMessage := "error finding refresh token"
		log.Printf("%s: %v", errorMessage, err)
		return "", service_errors.NewInternalServerError(errorMessage)
	}

	if refreshToken.ExpiresAt.Before(time.Now()) {
		errorMessage := "refresh token expired"
		log.Print(errorMessage)
		return "", service_errors.NewUnauthorizedError(errorMessage)
	}

	if refreshToken.RevokedAt.Valid {
		errorMessage := "refresh token revoked"
		log.Print(errorMessage)
		return "", service_errors.NewUnauthorizedError(errorMessage)
	}

	newAccessToken, err := auth.MakeJWT(refreshToken.UserID, s.Cfg.JWTSecret, constants.DEFAULT_EXPIRES_IN)
	if err != nil {
		errorMessage := "error generating access token"
		log.Printf("%s: %v", errorMessage, err)
		return "", service_errors.NewInternalServerError(errorMessage)
	}

	return newAccessToken, nil
}

func (s *authService) RevokeToken(ctx context.Context, input RevokeTokenInput) error {
	now := time.Now()
	err := s.Db.RevokeRefreshToken(ctx, database.RevokeRefreshTokenParams{
		Token: input.RefreshToken,
		RevokedAt: sql.NullTime{Time: now, Valid: true},
		UpdatedAt: now,
	})
	if err != nil {
		errorMessage := "error revoking refresh token"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	return nil
}

func (s *authService) UpgradeUser(ctx context.Context, input UpgradeUserInput) error {
	user, err := s.UserService.FindUserByID(ctx, FindUserByIDInput(input))
	if err != nil {
		errorMessage := "error finding user"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	_, err = s.Db.UpdateUserIsChirpyRed(ctx, database.UpdateUserIsChirpyRedParams{
		ID: user.ID,
		IsChirpyRed: true,
		UpdatedAt: time.Now(),
	})
	if err != nil {
		errorMessage := "error updating user is chirpy red"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	return nil
}
