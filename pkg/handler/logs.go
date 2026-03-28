package handler

import (
	"encoding/json"
	"net/http"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/rs/zerolog/log"
)

type logLine struct {
	Stream string `json:"stream"`
	Line   string `json:"line"`
}

type LogsHandler struct {
	client *client.Client
}

func NewLogsHandler(c *client.Client) *LogsHandler {
	return &LogsHandler{client: c}
}

func (h *LogsHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch, err := h.client.Logs(r.Context(), containerID, true)
	if err != nil {
		log.Error().Err(err).Str("container", containerID).Msg("logs failed")
		http.Error(w, "failed to get logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	enc := json.NewEncoder(w)
	for l := range ch {
		if l.Err != nil {
			log.Error().Err(l.Err).Str("container", containerID).Msg("logs stream error")
			return
		}
		if err := enc.Encode(logLine{Stream: l.Stream, Line: l.Line}); err != nil {
			return
		}
		flusher.Flush()
	}
}
