package handlers

import (
	"net/http"

	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/constants"
	"github.com/fekete965/boot.dev-chirpy-learn-http-server-in-go/internal/middlewares"
)

var HandleHome = middlewares.MiddlewareLogger(http.StripPrefix("/app", http.FileServer(http.Dir(constants.STATIC_DIR))))

var HandleAssets = middlewares.MiddlewareLogger(http.StripPrefix("/app/assets", http.FileServer(http.Dir(constants.STATIC_ASSETS_DIR))))
