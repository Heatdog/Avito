package bannerstransport

import (
	"log/slog"
	"net/http"

	banner_service "github.com/Heatdog/Avito/internal/service/bannerservice"
	"github.com/Heatdog/Avito/internal/transport"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	"github.com/gorilla/mux"
)

type bannersHandler struct {
	logger     *slog.Logger
	service    banner_service.BannerService
	middleware *middleware_transport.Middleware
}

func NewBannersHandler(logger *slog.Logger, service banner_service.BannerService,
	mid *middleware_transport.Middleware) transport.Handler {
	return &bannersHandler{
		logger:     logger,
		service:    service,
		middleware: mid,
	}
}

const (
	banner        = "/banner"
	userBanner    = "/user_banner"
	bannerID      = "/banner/{id}"
	bannerVersion = "/banner/{id}/{version}"
)

func (handler *bannersHandler) Register(router *mux.Router) {
	router.HandleFunc(banner, handler.middleware.Auth(handler.middleware.AdminAuth(handler.createBanner))).
		Methods(http.MethodPost)
	router.HandleFunc(userBanner, handler.middleware.Auth(handler.getUserBanner)).
		Methods(http.MethodGet)
	router.HandleFunc(banner, handler.middleware.Auth(handler.middleware.AdminAuth(handler.getBanners))).
		Methods(http.MethodGet)
	router.HandleFunc(bannerID, handler.middleware.Auth(handler.middleware.AdminAuth(handler.deleteBanner))).
		Methods(http.MethodDelete)
	router.HandleFunc(bannerID, handler.middleware.Auth(handler.middleware.AdminAuth(handler.updateBanner))).
		Methods(http.MethodPatch)
	router.HandleFunc(banner, handler.middleware.Auth(handler.middleware.AdminAuth(handler.deleteBannerOnTagOrFeature))).
		Methods(http.MethodDelete)
	router.HandleFunc(bannerVersion, handler.middleware.Auth(handler.middleware.AdminAuth(handler.updateBannerVersion))).
		Methods(http.MethodPatch)
}
