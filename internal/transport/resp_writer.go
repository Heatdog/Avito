package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type RespWriterError struct {
	Error string `json:"error"`
}

type RespWriterBannerCreated struct {
	BannerID int `json:"banner_id"`
}

func ResponseWriteError(w http.ResponseWriter, statusCode int, errString string, logger *slog.Logger) {
	w.WriteHeader(statusCode)
	w.Header().Add("content-type", "application/json")

	resp, err := json.Marshal(RespWriterError{
		Error: errString,
	})

	if err != nil {
		logger.Error(err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Debug("response error write", slog.String("err", errString), slog.Int("satus code", statusCode))
}

func ResponseWriteBannerCreated(w http.ResponseWriter, id int, logger *slog.Logger) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("content-type", "application/json")

	resp, err := json.Marshal(RespWriterBannerCreated{
		BannerID: id,
	})

	if err != nil {
		logger.Error(err.Error())
		return
	}

	if _, err := w.Write(resp); err != nil {
		logger.Error(err.Error())
		return
	}

	logger.Debug("response banner created", slog.Int("id", id))
}
