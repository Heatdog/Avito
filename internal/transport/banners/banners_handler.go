package banners_transport

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	banner_service "github.com/Heatdog/Avito/internal/service/banner"
	"github.com/Heatdog/Avito/internal/transport"
	middleware_transport "github.com/Heatdog/Avito/internal/transport/middleware"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
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
	router.HandleFunc(banner, handler.middleware.Auth(handler.middleware.AdminAuth(handler.getBanners))).
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
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}
	defer r.Body.Close()

	handler.logger.Debug("request body", slog.String("body", string(body)))

	handler.logger.Debug("unmarshaling request body")
	var banner banner_model.BannerInsert

	if err := json.Unmarshal(body, &banner); err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	handler.logger.Debug("validate request body", slog.Any("banner", banner))
	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("json", banner_model.ValidateJson)
	if err = validate.Struct(banner); err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}
	handler.logger.Debug("valid successful")

	id, err := handler.service.InsertBanner(r.Context(), banner)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	transport.ResponseWriteBannerCreated(w, id, handler.logger)
}

// Получение баннера для пользователя
// @Summary GetUserBanner
// @Security ApiKeyAuth
// @Description Получение баннера для пользователя
// @ID get-user-banner
// @Tags banner
// @Produce json
// @Param tag_id query integer true "tag_id"
// @Param feature_id query integer true "feature_id"
// @Param use_last_revision query boolean false "use_last_revision"
// @Success 200 {object} object JSON-отображение баннера
// @Failure 400 {object} transport.RespWriterError Некорректные данные
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 404 {object} nil Баннер не найден
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /user_banner [get]
func (handler *bannersHandler) getUserBanner(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("get user banner handler")

	handler.logger.Debug("read request query paarams")
	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	useLastRevisionStr := r.URL.Query().Get("use_last_revision")

	params, err := query_params.ValidateUserBannerParams(tagIdStr, featureIdStr, useLastRevisionStr)
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	handler.logger.Debug("params", slog.Int("tag_id", params.TagID), slog.Int("feature_id", params.FeatureID),
		slog.Bool("use_last_revision value", params.UseLastrRevision))

	content, err := handler.service.GetUserBanner(r.Context(), params)
	if err == pgx.ErrNoRows {
		handler.logger.Debug(err.Error())
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	resp, err := json.Marshal(content)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	if _, err = w.Write(resp); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}
	handler.logger.Debug(string(resp))
}

// Получение всех баннеров c фильтрацией по фиче и/или тегу
// @Summary GetBanners
// @Security ApiKeyAuth
// @Description Получение всех баннеров c фильтрацией по фиче и/или тегу
// @ID get-banner
// @Tags banner
// @Produce json
// @Param tag_id query integer false "tag_id"
// @Param feature_id query integer false "feature_id"
// @Param limit query integer false "limit"
// @Param offset query integer false "offset"
// @Success 200 {object} []banner_model.Banner Список баннеров
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /banner [get]
func (handler *bannersHandler) getBanners(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("get banners handler")

	handler.logger.Debug("read request query paarams")
	tagIdStr := r.URL.Query().Get("tag_id")
	featureIdStr := r.URL.Query().Get("feature_id")
	limitStr := r.URL.Query().Get("limit")
	offsetStr := r.URL.Query().Get("offset")

	params, err := query_params.ValidateBannersParams(tagIdStr, featureIdStr, limitStr, offsetStr)
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	handler.logger.Debug("params", slog.Any("params", params))

	banners, err := handler.service.GetBanners(r.Context(), params)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	resp, err := json.Marshal(banners)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")
	if _, err = w.Write(resp); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}
	handler.logger.Debug(string(resp))
}
