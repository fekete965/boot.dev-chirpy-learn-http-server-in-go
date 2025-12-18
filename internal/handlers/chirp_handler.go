package handlers

import (
	"fmt"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
	"github.com/google/uuid"
)

func (h *Handlers) HandleCreateChirp() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bearerToken, err := auth.GetBearerToken(r)
		if err != nil {
			errorMessage := fmt.Sprintf("error during authentication: %v", err)
			utils.RespondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}

		userID, err := auth.ValidateJWT(bearerToken, h.cfg.JWTSecret)
		if err != nil {
			errorMessage := fmt.Sprintf("error during authentication: %v", err)
			utils.RespondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}
	
		payload, err := utils.DecodeRequestBody[models.CreateChirpResource](r)
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		newChirp, err := h.services.ChirpService.CreateChirp(r.Context(), services.CreateChirpInput{
			UserID: userID,
			Body: payload.Body,
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		data := models.CreateChirpResponse{
			ID: newChirp.ID,
			UserID: newChirp.UserID,
			Body: newChirp.Body,
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
				errorMessage := fmt.Sprintf("invalid author_id: %v", err)
				utils.RespondWithJSON(w, http.StatusBadRequest, errorMessage)
				return
			}

			authorID = &parsedAuthorId
		}

		chirps, err := h.services.ChirpService.GetAllChirps(r.Context(), services.GetAllChirpsInput{
			UserID: authorID,
			Sort: sortParam,
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}
	
		data := make([]models.GetAllChirpResponse, len(chirps))
		for i, chirp := range chirps {
			data[i] = models.GetAllChirpResponse{
				ID: chirp.ID,
				UserID: chirp.UserID,
				Body: chirp.Body,
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
			errorMessage := fmt.Sprintf("cannot parse chirpID: %v", err)
			utils.RespondWithPlainText(w, http.StatusBadRequest, errorMessage)
			return
		}

		chirp, err := h.services.ChirpService.GetChirpByID(r.Context(), services.GetChirpByIDInput{
			ChirpID: chirpID,
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		data := models.GetChirpByIdResponse{
			ID: chirp.ID,
			UserId: chirp.UserID,
			Body: chirp.Body,
			CreatedAt: chirp.CreatedAt,
			UpdatedAt: chirp.UpdatedAt,
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}

func (h *Handlers) HandleDeleteChirp() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		accessToken, err := auth.GetBearerToken(r)
		if err != nil {
			errorMessage := fmt.Sprintf("error during authentication: %v", err)
			utils.RespondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}

		userID, err := auth.ValidateJWT(accessToken, h.cfg.JWTSecret)
		if err != nil {
			errorMessage := fmt.Sprintf("error validating token: %v", err)
			utils.RespondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}

		chirpID, err := utils.SafeParseUUID(r.PathValue("chirpID"))
		if err != nil {
			errorMessage := fmt.Sprintf("cannot parse chirpID: %v", err)
			utils.RespondWithPlainText(w, http.StatusBadRequest, errorMessage)
			return
		}

		err = h.services.ChirpService.DeleteChirp(r.Context(), services.DeleteChirpInput{
			UserID: userID,
			ChirpID: chirpID,
		})
		if err != nil {	
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		utils.RespondNoContent(w)
	}))
}
