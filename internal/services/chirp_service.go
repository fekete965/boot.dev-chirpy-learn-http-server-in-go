package services

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateChirpInput struct {
	UserID uuid.UUID
	Body string
}

type DeleteChirpInput struct {
	UserID uuid.UUID
	ChirpID uuid.UUID
}

type GetAllChirpsInput struct {
	UserID *uuid.UUID
	Sort *string
}

type GetChirpByIDInput struct {
	ChirpID uuid.UUID
}

type Chirp struct {
	ID uuid.UUID
	UserID uuid.UUID
	Body string
	CreatedAt time.Time
	UpdatedAt time.Time
}

type ChirpService interface {
	CreateChirp(ctx context.Context, input CreateChirpInput) (Chirp, error)
	DeleteChirp(ctx context.Context, input DeleteChirpInput) error
	GetAllChirps(ctx context.Context, input GetAllChirpsInput) ([]Chirp, error)
	GetChirpByID(ctx context.Context, input GetChirpByIDInput) (Chirp, error)
}
