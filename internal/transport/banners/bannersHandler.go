package banners_transport

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	"github.com/Heatdog/Avito/internal/transport"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	"github.com/go-playground/validator"
	"github.com/gorilla/mux"
)

type bannersHandler struct {
	logger     *slog.Logger
	service    banner_service.BannerService
	middleware *middleware_transport.Middleware
}

func NewBunnersHandler(logger *slog.Logger, service banner_service.BannerService,
	mid *middleware_transport.Middleware) transport.Handler {
	return &bannersHandler{
		logger:     logger,
		service:    service,
		middleware: mid,
	}
}

const (
	banner     = "/banner"
	userBanner = "/user_banner"
)

func (handler *bannersHandler) Register(router *mux.Router) {
	router.HandleFunc(banner, handler.middleware.Auth(handler.middleware.AdminAuth(handler.createBanner))).
		Methods(http.MethodPost)
	router.HandleFunc(userBanner, handler.middleware.Auth(handler.getUserBanner)).
		Methods(http.MethodGet)
}

// Создание нового баннера
// @Summary CreateBanner
// @Security ApiKeyAuth
// @Description Создание нового баннера
// @ID create-banner
// @Tags banner
// @Accept json
// @Produce json
// @Param input body banner_model.BannerInsert true "banner info"
// @Success 201 {object} transport.RespWriterBannerCreated ID созданного баннера
// @Failure 400 {object} transport.RespWriterError Некорректные данные
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /banner [post]
func (handler *bannersHandler) createBanner(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("create banner handler")

	handler.logger.Debug("read request body")
	body, err := io.ReadAll(r.Body)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}
	defer r.Body.Close()

	handler.logger.Debug("request body", slog.String("body", string(body)))

	handler.logger.Debug("unmarshaling request body")
	var banner banner_model.BannerInsert

	if err := json.Unmarshal(body, &banner); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	handler.logger.Debug("validate request body")
	validate := validator.New()
	if err = validate.Struct(banner); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	id, err := handler.service.InsertBanner(r.Context(), banner)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	transport.ResponseWriteBannerCreated(w, id, handler.logger)
}

func (handler *bannersHandler) getUserBanner(w http.ResponseWriter, r *http.Request) {

}
