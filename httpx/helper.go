package httpx

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
)

func writeJSON(w http.ResponseWriter, status int, payload interface{}) {
	body, err := json.Marshal(payload)
	if err != nil {
		slog.Default().Error(
			"Failed to marshal JSON response",
			slog.Int("status", status),
			slog.String("payload_type", typeNameOf(payload)),
			slog.String("error", err.Error()),
		)
		http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(body); err != nil {
		slog.Default().Error(
			"Failed to write JSON response",
			slog.Int("status", status),
			slog.Int("bytes", len(body)),
			slog.String("error", err.Error()),
		)
		return
	}
}

func typeNameOf(v any) string {
	if v == nil {
		return "<nil>"
	}
	return fmt.Sprintf("%T", v)
}
