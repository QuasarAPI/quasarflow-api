package middleware

import (
	"fmt"
	"net/http"
	"runtime/debug"

	"quasarflow-api/internal/interface/http/response"

	"go.uber.org/zap"
)

// Recovery is a middleware that recovers from panics and returns a 500 error
func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				// Log the panic
				logger, _ := zap.NewProduction()
				defer logger.Sync()

				logger.Error("Panic recovered",
					zap.String("error", fmt.Sprintf("%v", err)),
					zap.String("stack", string(debug.Stack())),
					zap.String("method", r.Method),
					zap.String("path", r.URL.Path),
				)

				// Return 500 error
				response.Error(w, http.StatusInternalServerError, "Internal server error")
			}
		}()

		next.ServeHTTP(w, r)
	})
}
