package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateChirpResource struct {
	Body string `json:"body" validate:"required"`
}

type CreateChirpResponse struct {
	ID uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetChirpByIdResponse struct {
	ID uuid.UUID `json:"id"`
	UserId uuid.UUID `json:"user_id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type GetAllChirpResponse struct {
	ID uuid.UUID `json:"id"`
	UserID uuid.UUID `json:"user_id"`
	Body string `json:"body"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
