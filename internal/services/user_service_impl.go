package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/google/uuid"
	"github.com/lib/pq"
)


type userService struct {
	Db *database.Queries
}

func NewUserService(Db *database.Queries) *userService {
	return &userService{Db: Db}
}

func (s *userService) CreateUser(ctx context.Context, input CreateUserInput) (User, error) {
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		errorMessage := "error hashing password"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	now := time.Now()
	newUser, err := s.Db.CreateUser(ctx, database.CreateUserParams{
		ID: uuid.New(),
		Email: input.Email,
		HashedPassword: hashedPassword,
		CreatedAt: now,
		UpdatedAt: now,
	})

	if err != nil {
		var pqErr *pq.Error

		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			errorMessage := "email already exists"
			log.Print(errorMessage)
			return User{}, service_errors.NewConflictError(errorMessage)
		}

		errorMessage := "user signup failed"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	return User{
		ID: newUser.ID,
		Email: newUser.Email,
		HashedPassword: newUser.HashedPassword,
		IsChirpyRed: newUser.IsChirpyRed,
		CreatedAt: newUser.CreatedAt,
		UpdatedAt: newUser.UpdatedAt,
	}, nil
}

func (s *userService) DeleteAllUsers(ctx context.Context) error {
	err := s.Db.DeleteAllUsers(ctx)
	if err != nil {
		errorMessage := "failed to delete all users"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	return nil
}

func (s *userService) FindUserByEmail(ctx context.Context, input FindUserByEmailInput) (User, error) {
	user, err := s.Db.FindUserByEmail(ctx, input.Email)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			errorMessage := "user not found"
			log.Print(errorMessage)
			return User{}, service_errors.NewNotFoundError(errorMessage)
		}


		errorMessage := "failed to find user by email"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	return User{
		ID: user.ID,
		Email: user.Email,
		HashedPassword: user.HashedPassword,
		IsChirpyRed: user.IsChirpyRed,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) FindUserByID(ctx context.Context, input FindUserByIDInput) (User, error) {
	user, err := s.Db.FindUserById(ctx, input.UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {

			errorMessage := "user not found"
			log.Print(errorMessage)
			return User{}, service_errors.NewNotFoundError(errorMessage)
		}


		errorMessage := "failed to find user by id"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	return User{
		ID: user.ID,
		Email: user.Email,
		HashedPassword: user.HashedPassword,
		IsChirpyRed: user.IsChirpyRed,
		CreatedAt: user.CreatedAt,
		UpdatedAt: user.UpdatedAt,
	}, nil
}

func (s *userService) UpdateUser(ctx context.Context, input UpdateUserInput) (User, error) {
	hashedPassword, err := auth.HashPassword(input.Password)
	if err != nil {
		errorMessage := "error hashing password"
		log.Printf("%s: %v", errorMessage, err)

		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	updatedUser, err := s.Db.UpdateUser(ctx, database.UpdateUserParams{
		ID: input.UserID,
		Email: input.Email,
		HashedPassword: hashedPassword,
	})
	if err != nil {
		errorMessage := "user update failed"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	return User{
		ID: updatedUser.ID,
		Email: updatedUser.Email,
		HashedPassword: updatedUser.HashedPassword,
		IsChirpyRed: updatedUser.IsChirpyRed,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}

func (s *userService) UpdateUserIsChirpyRed(ctx context.Context, input UpdateUserIsChirpyRedInput) (User, error) {
	updatedUser, err := s.Db.UpdateUserIsChirpyRed(ctx, database.UpdateUserIsChirpyRedParams{
		ID: input.UserID,
		IsChirpyRed: input.IsChirpyRed,
		UpdatedAt: input.UpdatedAt,
	})
	if err != nil {
		errorMessage := "user update failed"
		log.Printf("%s: %v", errorMessage, err)
		return User{}, service_errors.NewInternalServerError(errorMessage)
	}

	return User{
		ID: updatedUser.ID,
		Email: updatedUser.Email,
		HashedPassword: updatedUser.HashedPassword,
		IsChirpyRed: updatedUser.IsChirpyRed,
		CreatedAt: updatedUser.CreatedAt,
		UpdatedAt: updatedUser.UpdatedAt,
	}, nil
}
