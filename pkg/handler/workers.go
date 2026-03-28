package handler

import (
	"encoding/json"
	"maps"
	"net/http"
	"slices"

	"agentsmanager/pkg/config"
)

type WorkersHandler struct {
	store *config.Store
}

func NewWorkersHandler(store *config.Store) *WorkersHandler {
	return &WorkersHandler{store: store}
}

func (h *WorkersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	names := slices.Sorted(maps.Keys(h.store.Workers()))
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(names)
}
