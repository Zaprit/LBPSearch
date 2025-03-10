package utils

import (
	"log/slog"
	"net/http"
)

func HttpLog(w http.ResponseWriter, status int, message string) {
	w.WriteHeader(status)
	_, err := w.Write([]byte(message))
	slog.Debug("HTTP Log", "status", status, "message", message)
	if err != nil {
		slog.Error("failed to write to ResponseWriter")
	}
}
