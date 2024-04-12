package banners_transport

import (
	"encoding/json"
	"io"
	"log/slog"
	"net/http"

	banner_model "github.com/Heatdog/Avito/internal/models/banner"
	"github.com/Heatdog/Avito/internal/transport"
	"github.com/go-playground/validator/v10"
)

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

	if err = validate.RegisterValidation("json", banner_model.ValidateJson); err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}

	if err = validate.Struct(banner); err != nil {
		handler.logger.Debug(err.Error())
		transport.ResponseWriteError(w, http.StatusBadRequest, err.Error(), handler.logger)
		return
	}
	handler.logger.Debug("valid successful")

	id, err := handler.service.InsertBanner(r.Context(), &banner)
	if err != nil {
		handler.logger.Warn(err.Error())
		transport.ResponseWriteError(w, http.StatusInternalServerError, err.Error(), handler.logger)
		return
	}

	transport.ResponseWriteBannerCreated(w, id, handler.logger)
}
