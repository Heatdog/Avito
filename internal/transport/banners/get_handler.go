package banners_transport

import (
	"encoding/json"
	"log/slog"
	"net/http"

	_ "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/models/query_params"
	"github.com/Heatdog/Avito/internal/transport"
	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgx/v5"
)

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

	token := r.Context().Value("token").(string)
	handler.logger.Debug("token header", slog.String("token", token))

	params := query_params.BannerUserParams{
		TagID:            r.URL.Query().Get("tag_id"),
		FeatureID:        r.URL.Query().Get("feature_id"),
		UseLastrRevision: r.URL.Query().Get("use_last_revision"),
		Token:            token,
	}

	handler.logger.Debug("validate request params", slog.Any("params", params))
	validate := validator.New(validator.WithRequiredStructEnabled())
	if err := validate.Struct(params); err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}
	handler.logger.Debug("valid successful")

	content, err := handler.service.GetUserBanner(r.Context(), &params)
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

	handler.logger.Debug("read request query params")
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

	banners, err := handler.service.GetBanners(r.Context(), &params)
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
