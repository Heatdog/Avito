package bannerstransport

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/transport"
	"github.com/go-playground/validator/v10"
	"github.com/gorilla/mux"
	"github.com/jackc/pgx/v5"
)

// Обновление содержимого баннера
// @Summary UpdateBanner
// @Security ApiKeyAuth
// @Description Обновление содержимого баннера
// @ID update-banner
// @Tags banner
// @Produce json
// @Param id path integer true "id"
// @Param input body banner_model.BannerUpdate true "banner info"
// @Success 200 {object} nil OK
// @Failure 400 {object} transport.RespWriterError Некорректные данные
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 404 {object} nil Баннер не найден
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /banner/{id} [patch]
func (handler *bannersHandler) updateBanner(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}
	defer r.Body.Close()

	handler.logger.Debug("update banner handler", slog.Int("id", id), slog.String("body", string(body)))
	handler.logger.Debug("unmarshaling request body")

	var banner banner_model.BannerUpdate

	if err := json.Unmarshal(body, &banner); err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	banner.ID = id

	handler.logger.Debug("validate request body", slog.Any("banner", banner))

	validate := validator.New(validator.WithRequiredStructEnabled())

	if err = validate.RegisterValidation("json", banner_model.ValidateJSON); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	if err = validate.Struct(banner); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	handler.logger.Debug("valid successful")

	err = handler.service.UpdateBanner(r.Context(), &banner)
	if err == pgx.ErrNoRows {
		handler.logger.Debug(err.Error())
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)

		return
	}

	w.WriteHeader(http.StatusOK)
	handler.logger.Debug("update OK")
}

// Обновление последней версии баннера
// @Summary UpdateBannerVersion
// @Security ApiKeyAuth
// @Description Обновление последней версии баннера
// @ID update-banner-version
// @Tags banner
// @Produce json
// @Param id path integer true "id"
// @Param version path integer true "version"
// @Success 200 {object} nil OK
// @Failure 400 {object} transport.RespWriterError Некорректные данные
// @Failure 401 {object} nil Пользователь не авторизован
// @Failure 403 {object} nil Пользователь не имеет доступа
// @Failure 404 {object} nil Баннер не найден
// @Failure 500 {object} transport.RespWriterError Внутренняя ошибка сервера
// @Router /banner/{id}/{version} [patch]
func (handler *bannersHandler) updateBannerVersion(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(mux.Vars(r)["id"])
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	version, err := strconv.Atoi(mux.Vars(r)["version"])
	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	if version < 1 || version > 3 {
		err = fmt.Errorf("bad version")
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)

		return
	}

	handler.logger.Debug("update banner handler", slog.Int("id", id), slog.Int("version", version))

	err = handler.service.UpdateBannerVersion(r.Context(), id, version)
	if err == pgx.ErrNoRows {
		handler.logger.Debug(err.Error())
		w.WriteHeader(http.StatusNotFound)

		return
	}

	if err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)

		return
	}

	w.WriteHeader(http.StatusOK)
	handler.logger.Debug("update OK")
}
