package handlers

import (
	"net/http"
	"time"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/auth"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/models"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/services"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
)

func (h *Handlers) HandleReset() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		if h.Cfg.Platform != "dev" {
			utils.RespondWithPlainText(w, http.StatusForbidden, "Forbidden operation")
			return
		}

		// Reset the metrics
		h.Cfg.FileserverHits.Store(0)

		// Reset the database
		err := h.Services.UserService.DeleteAllUsers(r.Context())
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		w.Header().Add("Content-Type", "text/plain; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Metric has been reset"))
	}))
}

func (h *Handlers) HandlePolkaWebhooks() http.Handler {
	return middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiKey, err := auth.GetAPIKey(r)
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		if apiKey != h.Cfg.PolkaWebhookSecret {
			errorMessage := "invalid api key"
			utils.RespondWithPlainText(w, http.StatusUnauthorized, errorMessage)
			return
		}
		
		payload, err := utils.DecodeRequestBody[models.WebhookResource](r)
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		if payload.Event != "user.upgraded" {
			utils.RespondWithNoContent(w)
			return
		}

		user, err := h.Services.UserService.FindUserByID(r.Context(), services.FindUserByIDInput{
			UserID: payload.Data.UserID,
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		_, err = h.Services.UserService.UpdateUserIsChirpyRed(r.Context(), services.UpdateUserIsChirpyRedInput{
			UserID: user.ID,
			IsChirpyRed: true,
			UpdatedAt: time.Now(),
		})
		if err != nil {
			statusCode, errorMessage := utils.ServiceErrorToRequestError(err)
			utils.RespondWithPlainText(w, statusCode, errorMessage)
			return
		}

		utils.RespondWithNoContent(w)
	}))
}
