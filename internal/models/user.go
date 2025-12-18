package models

import (
	"time"

	"github.com/google/uuid"
)

type CreateUserResource struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type CreateUserResponse struct {
	Id uuid.UUID `json:"id"`
	Email string `json:"email"`
	IsChirpyRed bool `json:"is_chirpy_red"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type UpdateUserResource struct {
	Email string `json:"email"`
	Password string `json:"password"`
}

type UpdateUserResponse struct {
	Id uuid.UUID `json:"id"`
	Email string `json:"email"`
	IsChirpyRed bool `json:"is_chirpy_red"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
