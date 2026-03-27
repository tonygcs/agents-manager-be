package handler

import (
	"encoding/json"
	"net/http"
)

type WorkerType string

const (
	WorkerTypeOpencodeNode WorkerType = "opencode-node"
	WorkerTypeCountdown    WorkerType = "countdown"
)

var workerTypes = []WorkerType{
	WorkerTypeOpencodeNode,
	WorkerTypeCountdown,
}

type WorkersHandler struct{}

func NewWorkersHandler() *WorkersHandler {
	return &WorkersHandler{}
}

func (h *WorkersHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(workerTypes)
}
