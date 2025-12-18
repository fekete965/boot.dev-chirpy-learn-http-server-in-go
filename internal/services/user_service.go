package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateUserInput struct {
	Email string
	Password string
}

type FindUserByEmailInput struct {
	Email string
}

type FindUserByIDInput struct {
	UserID uuid.UUID
}

type UpdateUserInput struct {
	UserID uuid.UUID
	Email string
	Password string
}

type UpdateUserIsChirpyRedInput struct {
	UserID uuid.UUID
	IsChirpyRed bool
	UpdatedAt time.Time
}

type User struct {
	ID uuid.UUID
	Email string
	HashedPassword string
	IsChirpyRed bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

type UserService interface {
	CreateUser(ctx context.Context, input CreateUserInput) (User, error)
	DeleteAllUsers(ctx context.Context) error
	FindUserByEmail(ctx context.Context, input FindUserByEmailInput) (User, error)
	FindUserByID(ctx context.Context, input FindUserByIDInput) (User, error)
	UpdateUser(ctx context.Context, input UpdateUserInput) (User, error)
	UpdateUserIsChirpyRed(ctx context.Context, input UpdateUserIsChirpyRedInput) (User, error)
}
