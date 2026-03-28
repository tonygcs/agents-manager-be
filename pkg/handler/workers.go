package handler

import (
	"encoding/json"
	"net/http"
)

type WorkersHandler struct {
	names []string
}

func NewWorkersHandler(workers []string) *WorkersHandler {
	return &WorkersHandler{names: workers}
}

func (h *WorkersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(h.names)
}
