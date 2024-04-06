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
	BannerId int `json:"banner_id"`
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

	logger.Debug("response error write", slog.String("err", errString), slog.Int("sttus code", statusCode))
}

func ResponseWriteBannerCreated(w http.ResponseWriter, id int, logger *slog.Logger) {
	w.WriteHeader(http.StatusCreated)
	w.Header().Add("content-type", "application/json")

	resp, err := json.Marshal(RespWriterBannerCreated{
		BannerId: id,
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

func ResponseWriteBanner(w http.ResponseWriter, content string, logger *slog.Logger) {
	w.WriteHeader(http.StatusOK)
	w.Header().Add("content-type", "application/json")

	if _, err := w.Write([]byte(content)); err != nil {
		logger.Error(err.Error())
		return
	}
}
