package services

import (
	"context"
	"database/sql"
	"errors"
	"log"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/config"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/database"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
	"github.com/google/uuid"
)

type NewChirpServiceInput struct {
	cfg *config.ApiConfig
	db *database.Queries
	userService UserService
}

type chirpService struct {
	cfg *config.ApiConfig
	db *database.Queries
	userService UserService
}

func NewChirpService(input NewChirpServiceInput) *chirpService {
	return &chirpService{
		cfg: input.cfg,
		db: input.db,
		userService: input.userService,
	}
}

func (s *chirpService) CreateChirp(ctx context.Context, input CreateChirpInput) (Chirp, error) {
	err := utils.ValidateChirpLength(input.Body, constants.MAX_CHIRP_LENGTH)
	if err != nil {
		errorMessage := "invalid chirp length"
		log.Printf("%s: %v", errorMessage, err)
		return Chirp{}, service_errors.NewBadRequestError(errorMessage)
	}

	user, err := s.userService.FindUserByID(ctx, FindUserByIDInput{UserID: input.UserID})
	if err != nil {
		errorMessage := "failed to find user"
		log.Printf("%s: %v", errorMessage, err)
		return Chirp{}, service_errors.NewInternalServerError(errorMessage)
	}


	cleanedBody := utils.CleanChirp(input.Body, constants.PROFANE_WORDS)

	newChirp, err := s.db.CreateChirp(ctx, database.CreateChirpParams{
		ID: uuid.New(),
		UserID: user.ID,
		Body: cleanedBody,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})		

	if err != nil {
		errorMessage := "failed to create chirp"
		log.Printf("%s: %v", errorMessage, err)
		return Chirp{}, service_errors.NewInternalServerError(errorMessage)
	}

	return Chirp{
		ID: newChirp.ID,
		UserID: newChirp.UserID,
		Body: newChirp.Body,
		CreatedAt: newChirp.CreatedAt,
		UpdatedAt: newChirp.UpdatedAt,
	}, nil
}

func (s *chirpService) DeleteChirp(ctx context.Context, input DeleteChirpInput) error {
	user, err := s.userService.FindUserByID(ctx, FindUserByIDInput{UserID: input.UserID})
	if err != nil {
		errorMessage := "failed to find user"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	chirp, err := s.GetChirpByID(ctx, GetChirpByIDInput{ChirpID: input.ChirpID})
	if err != nil {
		errorMessage := "failed to get chirp by id"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	if chirp.UserID != user.ID {
		errorMessage := "invalid user permission"
		log.Print(errorMessage)
		return service_errors.NewForbiddenError(errorMessage)
	}

	err = s.db.DeleteChirp(ctx, database.DeleteChirpParams{
		ID: chirp.ID,
		UserID: user.ID,
	})
	if err != nil {
		errorMessage := "failed to delete chirp"
		log.Printf("%s: %v", errorMessage, err)
		return service_errors.NewInternalServerError(errorMessage)
	}

	return nil
}

func (s *chirpService) GetAllChirps(ctx context.Context, input GetAllChirpsInput) ([]Chirp, error) {
	chirps, err := s.db.GetAllChirps(ctx, database.GetAllChirpsParams{
		AuthorID: input.UserID,
		Sort: input.Sort,
	})
	if err != nil {
		errorMessage := "failed to get all chirps"
		log.Printf("%s: %v", errorMessage, err)
		return nil, service_errors.NewInternalServerError(errorMessage)
	}

	data := make([]Chirp, len(chirps))
	for i, chirp := range chirps {
		data[i] = Chirp{
			ID: chirp.ID,
			UserID: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}
	}

	return data, nil
}

func (s *chirpService) GetChirpByID(ctx context.Context, input GetChirpByIDInput) (Chirp, error) {
	chirp, err := s.db.GetChirpById(ctx, input.ChirpID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			errorMessage := "chirp not found"
			log.Print(errorMessage)
			return Chirp{}, service_errors.NewNotFoundError(errorMessage)
		}

		errorMessage := "failed to get chirp by id"
		log.Printf("%s: %v", errorMessage, err)
		return Chirp{}, service_errors.NewInternalServerError(errorMessage)
	}

	return Chirp{
		ID: chirp.ID,
		UserID: chirp.UserID,
		Body: chirp.Body,
		CreatedAt: chirp.CreatedAt,
		UpdatedAt: chirp.UpdatedAt,
	}, nil
}
