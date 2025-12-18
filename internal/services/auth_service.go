package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type LoginInput struct {
	Email string
	Password string
}

type LoginOutput struct {
	UserID uuid.UUID
	Email string
	IsChirpyRed bool
	CreatedAt time.Time
	UpdatedAt time.Time
	Token string
	RefreshToken string
}

type RefreshTokenInput struct {
	RefreshToken string
}

type RevokeTokenInput struct {
	RefreshToken string
}

type UpgradeUserInput struct {
	UserID uuid.UUID
}

type AuthService interface {
	Login(ctx context.Context, input LoginInput) (LoginOutput, error) 
	RefreshToken(ctx context.Context, input RefreshTokenInput) (string, error)
	RevokeToken(ctx context.Context, input RevokeTokenInput) error
	UpgradeUser(ctx context.Context, input UpgradeUserInput) error
}
