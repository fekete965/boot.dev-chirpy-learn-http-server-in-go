package handlers

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/utils"
)

var HandleHealthCheck http.Handler = middlewares.MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	utils.RespondWithPlainText(w, http.StatusOK, "OK")
}))
