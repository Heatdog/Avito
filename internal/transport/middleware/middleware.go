package middleware_transport

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/Heatdog/Avito/internal/transport"
	"github.com/Heatdog/Avito/pkg/token"
)

type Middleware struct {
	logger        *slog.Logger
	tokenProvider token.TokenProvider
}

func NewMiddleware(logger *slog.Logger, tokenProvider token.TokenProvider) *Middleware {
	return &Middleware{
		logger:        logger,
		tokenProvider: tokenProvider,
	}
}

func (mid *Middleware) Auth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token := r.Header.Get("token")

		mid.logger.Debug("verify token", slog.String("token", token))

		if token == "" {
			mid.logger.Debug("token is empty")
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		if !mid.tokenProvider.VerifyToken(token) {
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
		if !mid.tokenProvider.VerifyOnAdmin(token.(string)) {
			mid.logger.Debug("not admin token")
			w.WriteHeader(http.StatusForbidden)
			return
		}

		next(w, r)
	}
}

func (mid *Middleware) Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			err := recover()
			if err != nil {
				transport.ResponseWriteError(w, http.StatusInternalServerError, err.(error).Error(), mid.logger)
			}
		}()

		mid.logger.Debug("request",
			slog.String("URL", r.URL.Path),
			slog.String("method", r.Method),
			slog.String("host", r.RemoteAddr))

		next.ServeHTTP(w, r)
	})
}
