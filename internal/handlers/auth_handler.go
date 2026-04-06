package handlers

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
)

func (h *Handlers) HandleLogin() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		payload, err := utils.DecodeRequestBody[models.LoginResource](r)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		loginOutput, err := h.Services.AuthService.Login(r.Context(), services.LoginInput{
			Email:    payload.Email,
			Password: payload.Password,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		data := models.LoginResponse{
			ID:           loginOutput.UserID,
			Email:        loginOutput.Email,
			IsChirpyRed:  loginOutput.IsChirpyRed,
			CreatedAt:    loginOutput.CreatedAt,
			UpdatedAt:    loginOutput.UpdatedAt,
			Token:        loginOutput.Token,
			RefreshToken: loginOutput.RefreshToken,
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}

func (h *Handlers) HandleTokenRefresh() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := utils.GetBearerToken(r)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		newAccessToken, err := h.Services.AuthService.RefreshToken(r.Context(), services.RefreshTokenInput{
			RefreshToken: token,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		data := models.TokenRefreshResponse{
			Token: newAccessToken,
		}

		utils.RespondWithJSON(w, http.StatusOK, data)
	}))
}

func (h *Handlers) HandleTokenRevoke() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		token, err := utils.GetBearerToken(r)
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		err = h.Services.AuthService.RevokeToken(r.Context(), services.RevokeTokenInput{
			RefreshToken: token,
		})
		if err != nil {
			statusCode, apiErrorResponse := utils.ServiceErrorToApiResponse(err)
			utils.RespondWithJSON(w, statusCode, apiErrorResponse)
			return
		}

		utils.RespondWithNoContent(w)
	}))
}
