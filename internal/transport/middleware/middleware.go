package middleware_transport

import (
	"context"
	"log/slog"
	"net/http"
)

type Middleware struct {
	logger *slog.Logger
}

func NewMiddleware(logger *slog.Logger) *Middleware {
	return &Middleware{
		logger: logger,
	}
}

const (
	userToken  = "user_token"
	adminToken = "admin_token"
)

func (mid *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		mid.logger.Debug("verify token", slog.String("token", token))

		if token == "" {
			mid.logger.Debug("token is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !mid.verifyToken(token) {
			mid.logger.Debug("token incorrect")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		ctx := context.WithValue(r.Context(), "token", token)
		next(w, r.WithContext(ctx))
	}
}

func (mid *Middleware) AdminAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Context().Value("token")
		if token == nil {
			mid.logger.Debug("empty token")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		mid.logger.Debug("token", token)
		if token.(string) != adminToken {
			mid.logger.Debug("not admin token")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func (mid *Middleware) verifyToken(token string) bool {
	return token == userToken || token == adminToken
}
