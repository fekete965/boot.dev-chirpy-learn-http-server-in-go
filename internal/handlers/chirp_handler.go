package handlers

import (
	"fmt"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/service_errors"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
	"github.com/google/uuid"
)

func (h *Handlers) HandleCreateChirp() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _, err := utils.GetAuthenticatedUserID(r, h.Cfg.JWTSecret)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		payload, err := utils.DecodeRequestBody[models.CreateChirpResource](r)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		newChirp, err := h.Services.ChirpService.CreateChirp(r.Context(), services.CreateChirpInput{
			UserID: userID,
			Body:   payload.Body,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		data := models.CreateChirpResponse{
			ID:        newChirp.ID,
			UserID:    newChirp.UserID,
			Body:      newChirp.Body,
			CreatedAt: newChirp.CreatedAt,
			UpdatedAt: newChirp.UpdatedAt,
		}

		utils.RespondWithJSON(w, http.StatusCreated, data)
	}))
}

func (h *Handlers) HandleGetAllChirps() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		authorIdParam := r.URL.Query().Get("author_id")
		sortParam := utils.GetQueryParam(r, "sort", &constants.DEFAULT_SORT)

		var authorID *uuid.UUID

		if authorIdParam != "" {
			parsedAuthorId, err := utils.SafeParseUUID(authorIdParam)

			if err != nil {
				parseError := service_errors.NewBadRequestError(fmt.Sprintf("invalid author_id: %v", err))
				statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(parseError)
				utils.RespondWithJSON(w, statusCode, apiErrorResponse)
				return
			}

			authorID = &parsedAuthorId
		}

		chirps, err := h.Services.ChirpService.GetAllChirps(r.Context(), services.GetAllChirpsInput{
			UserID: authorID,
			Sort:   sortParam,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		data := make([]models.GetAllChirpResponse, len(chirps))
		for i, chirp := range chirps {
			data[i] = models.GetAllChirpResponse{
				ID:        chirp.ID,
				UserID:    chirp.UserID,
				Body:      chirp.Body,
				CreatedAt: chirp.CreatedAt,
				UpdatedAt: chirp.UpdatedAt,
			}
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}

func (h *Handlers) HandleGetChirpById() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		chirpID, err := utils.SafeParseUUID(r.PathValue("chirpID"))
		if err != nil {
			parseError := service_errors.NewBadRequestError(fmt.Sprintf("cannot parse chirpID: %v", err))
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(parseError)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		chirp, err := h.Services.ChirpService.GetChirpByID(r.Context(), services.GetChirpByIDInput{
			ChirpID: chirpID,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		data := models.GetChirpByIdResponse{
			ID:        chirp.ID,
			UserID:    chirp.UserID,
			Body:      chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}

func (h *Handlers) HandleDeleteChirp() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		userID, _, err := utils.GetAuthenticatedUserID(r, h.Cfg.JWTSecret)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		chirpID, err := utils.SafeParseUUID(r.PathValue("chirpID"))
		if err != nil {
			parseError := service_errors.NewBadRequestError(fmt.Sprintf("cannot parse chirpID: %v", err))
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(parseError)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		err = h.Services.ChirpService.DeleteChirp(r.Context(), services.DeleteChirpInput{
			UserID:  userID,
			ChirpID: chirpID,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		utils.RespondWithNoContent(w)
	}))
}
