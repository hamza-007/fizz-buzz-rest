package rest

import (
	"log/slog"
	"net/http"

	resp "fizz-buzz-rest/utils/resp"
)

func (a *API) Health(w http.ResponseWriter, r *http.Request) {
	if err := a.health(r.Context()); err != nil {
		slog.ErrorContext(r.Context(), "health check failed", "error", err)
		resp.SendStatus(r.Context(), w, http.StatusServiceUnavailable, "unavailable", "service is not ready")
		return
	}
	resp.SendStruct(r.Context(), w, http.StatusOK, map[string]string{"status": "ok"})
}
