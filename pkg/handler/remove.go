package handler

import (
	"log"
	"net/http"

	"github.com/algorath-software/workerd/pkg/client"
)

type RemoveHandler struct {
	client *client.Client
}

func NewRemoveHandler(c *client.Client) *RemoveHandler {
	return &RemoveHandler{client: c}
}

func (h *RemoveHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	containerID := r.PathValue("id")

	if err := h.client.Stop(r.Context(), containerID, 5); err != nil {
		log.Printf("stop failed for container %s: %v", containerID, err)
	}

	if err := h.client.Remove(r.Context(), containerID, true); err != nil {
		log.Printf("remove failed for container %s: %v", containerID, err)
		http.Error(w, "failed to remove container", http.StatusInternalServerError)
		return
	}

	log.Printf("removed container %s", containerID)
	w.WriteHeader(http.StatusNoContent)
}
