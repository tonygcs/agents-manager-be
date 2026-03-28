package handler

import (
	"encoding/json"
	"net/http"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/rs/zerolog/log"
)

type containerItem struct {
	ID     string            `json:"id"`
	Name   string            `json:"name"`
	Image  string            `json:"image"`
	Status string            `json:"status"`
	Labels map[string]string `json:"labels"`
}

type ContainersHandler struct {
	client *client.Client
}

func NewContainersHandler(c *client.Client) *ContainersHandler {
	return &ContainersHandler{client: c}
}

func (h *ContainersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	containers, err := h.client.List(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("list containers failed")
		http.Error(w, "failed to list containers", http.StatusInternalServerError)
		return
	}

	var list []containerItem
	for _, c := range containers {
		list = append(list, containerItem{ID: c.ID, Name: c.Name, Image: c.Image, Status: c.Status, Labels: c.Labels})
	}
	if list == nil {
		list = []containerItem{}
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(list)
}

type ContainerHandler struct {
	client *client.Client
}

func NewContainerHandler(c *client.Client) *ContainerHandler {
	return &ContainerHandler{client: c}
}

func (h *ContainerHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	containers, err := h.client.List(r.Context())
	if err != nil {
		log.Error().Err(err).Msg("list containers failed")
		http.Error(w, "failed to get container", http.StatusInternalServerError)
		return
	}

	for _, c := range containers {
		if c.ID == id {
			w.Header().Set("Content-Type", "application/json")
			json.NewEncoder(w).Encode(containerItem{ID: c.ID, Name: c.Name, Image: c.Image, Status: c.Status, Labels: c.Labels})
			return
		}
	}

	http.Error(w, "container not found", http.StatusNotFound)
}
