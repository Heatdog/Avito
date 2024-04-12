package bannerstransport

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/Heatdog/Avito/internal/models/queryparams"
	"github.com/Heatdog/Avito/internal/transport"
	"github.com/gorilla/mux"
)

// Удаление баннера по идентификатору
// @Summary DeleteBanner
// @Security ApiKeyAuth
// @Description Удаление баннера по идентификатору
// @ID delete-banner
// @Tags banner
// @Produce json
// @Param id path integer false "id"
// @Success 204 {object} nil Баннер успешно удален
// @Failure 400 {object} transport.RespWriterError Некорректные данные
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 404 {object} nil Баннер не найден
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /banner/{id} [delete]
func (handler *bannersHandler) deleteBanner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	handler.logger.Debug("delete banner handler", slog.Int("id", id))

	ok, err := handler.service.DeleteBanner(r.Context(), id)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)

		return
	}

	if !ok {
		handler.logger.Debug("banner not found")
		w.WriteHeader(http.StatusNotFound)

		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// Удаления баннеров по фиче или тегу
// @Summary DeleteBannerOnTagOrFeature
// @Security ApiKeyAuth
// @Description Удаления баннеров по фиче или тегу
// @ID delete-banner-tag-feature
// @Tags banner
// @Produce json
// @Param tag_id query integer false "tag_id"
// @Param feature_id query integer false "feature_id"
// @Success 202 {object} nil Принято
// @Router /banner [delete]
func (handler *bannersHandler) deleteBannerOnTagOrFeature(w http.ResponseWriter, r *http.Request) {
	handler.logger.Debug("read request query params")

	tagIDStr := r.URL.Query().Get("tag_id")
	featureIDStr := r.URL.Query().Get("feature_id")

	params, err := queryparams.ValidateDeleteBannerParams(tagIDStr, featureIDStr)
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	handler.logger.Debug("params", slog.Any("params", params))
	go handler.service.DeleteBanners(context.Background(), params)
	w.WriteHeader(http.StatusAccepted)
}
