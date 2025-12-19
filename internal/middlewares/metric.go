package middlewares

import (
	"fmt"
	"net/http"
)

func (h *Middlewares) MiddlewareMetricsInc(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.Cfg.FileserverHits.Add(1)
		next.ServeHTTP(w, r)	
	})
}

func (h *Middlewares) MiddlewareHandleMetrics() http.Handler {
	return MiddlewareLogger(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		result := fmt.Sprintf(`
			<html>
				<body>
					<h1>Welcome, Chirpy Admin</h1>
					<p>Chirpy has been visited %d times!</p>
				</body>
			</html>`,
			h.Cfg.FileserverHits.Load())
	
		w.Header().Add("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(result))
	}))
}
