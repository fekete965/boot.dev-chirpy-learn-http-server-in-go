package handlers

import (
	"fmt"
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
)

func (h *Handlers) HandleCreateUser() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := utils.DecodeRequestBody[models.CreateUserResource](r)
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		newUser, err := h.services.UserService.CreateUser(r.Context(), services.CreateUserInput{
			Email: payload.Email,
			Password: payload.Password,
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		data := models.CreateUserResponse{
			Id: newUser.ID,
			Email: newUser.Email,
			IsChirpyRed: newUser.IsChirpyRed,
			CreatedAt: newUser.CreatedAt,
			UpdatedAt: newUser.UpdatedAt,
		}

		utils.RespondWithJSON(w, http.StatusCreated, data)
	}))
}

func (h *Handlers) HandleUpdateUser() http.Handler {
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


		payload, err := utils.DecodeRequestBody[models.UpdateUserResource](r)
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
    	utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		updatedUser, err := h.services.UserService.UpdateUser(r.Context(), services.UpdateUserInput{
			UserID: userID,
			Email: payload.Email,
			Password: payload.Password,
		})
		
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}
		
		data := models.UpdateUserResponse{
			Id: updatedUser.ID,
			Email: updatedUser.Email,
			IsChirpyRed: updatedUser.IsChirpyRed,
			CreatedAt: updatedUser.CreatedAt,
			UpdatedAt: updatedUser.UpdatedAt,
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}
