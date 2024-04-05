package transport

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

type RespWriterError struct {
	Error string `json:"error"`
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
