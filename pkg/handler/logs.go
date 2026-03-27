package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/algorath-software/workerd/pkg/client"
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
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// path: /containers/{id}/logs
	parts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(parts) != 3 || parts[0] != "containers" || parts[2] != "logs" {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}
	containerID := parts[1]
	if containerID == "" {
		http.Error(w, "missing container id", http.StatusBadRequest)
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	ch, err := h.client.Logs(r.Context(), containerID, true)
	if err != nil {
		log.Printf("logs failed for container %s: %v", containerID, err)
		http.Error(w, "failed to get logs", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/x-ndjson")
	enc := json.NewEncoder(w)
	for l := range ch {
		if l.Err != nil {
			log.Printf("logs stream error for container %s: %v", containerID, l.Err)
			return
		}
		if err := enc.Encode(logLine{Stream: l.Stream, Line: l.Line}); err != nil {
			return
		}
		flusher.Flush()
	}
}
