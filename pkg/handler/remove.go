package handler

import (
	"net/http"
	"time"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/rs/zerolog/log"
)

type RemoveHandler struct {
	client *client.Client
}

func NewRemoveHandler(c *client.Client) *RemoveHandler {
	return &RemoveHandler{client: c}
}

func (h *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")

	if err := h.client.Stop(r.Context(), containerID, 60); err != nil {
		log.Error().Err(err).Str("container", containerID).Msg("stop failed")
	}

	if err := h.client.Remove(r.Context(), containerID, true); err != nil {
		log.Error().Err(err).Str("container", containerID).Msg("remove failed")
		http.Error(w, "failed to remove container", http.StatusInternalServerError)
		return
	}

	// Poll until the container is no longer listed.
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()
	timeout := time.After(30 * time.Second)
	for {
		select {
		case <-r.Context().Done():
			http.Error(w, "request cancelled", http.StatusServiceUnavailable)
			return
		case <-timeout:
			log.Error().Str("container", containerID).Msg("timed out waiting for container removal")
			http.Error(w, "timed out waiting for container removal", http.StatusInternalServerError)
			return
		case <-ticker.C:
			containers, err := h.client.List(r.Context())
			if err != nil {
				log.Error().Err(err).Str("container", containerID).Msg("list failed while polling removal")
				continue
			}
			found := false
			for _, c := range containers {
				if c.ID == containerID {
					found = true
					break
				}
			}
			if !found {
				log.Info().Str("container", containerID).Msg("removed")
				w.WriteHeader(http.StatusNoContent)
				return
			}
		}
	}
}
