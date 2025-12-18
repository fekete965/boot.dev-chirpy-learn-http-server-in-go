package middlewares

import (
	"log"
	"net/http"
)


func MiddlewareLogger(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%v: %v", r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}
