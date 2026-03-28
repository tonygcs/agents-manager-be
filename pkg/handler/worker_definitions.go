package handler

import (
	"encoding/json"
	"net/http"

	"agentsmanager/pkg/config"
	"github.com/rs/zerolog/log"
)

type workerConfigItem struct {
	Image   string            `json:"image"`
	Cmd     []string          `json:"cmd"`
	Labels  map[string]string `json:"labels"`
	Secrets []string          `json:"secrets"`
}

type WorkerDefinitionsHandler struct {
	store *config.Store
}

func NewWorkerDefinitionsHandler(store *config.Store) *WorkerDefinitionsHandler {
	return &WorkerDefinitionsHandler{store: store}
}

func (h *WorkerDefinitionsHandler) List(w http.ResponseWriter, r *http.Request) {
	workers := h.store.Workers()
	out := make(map[string]workerConfigItem, len(workers))
	for name, wc := range workers {
		out[name] = toWorkerConfigItem(wc)
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(out)
}

func (h *WorkerDefinitionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	wc, ok := h.store.Worker(name)
	if !ok {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(toWorkerConfigItem(wc))
}

func (h *WorkerDefinitionsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Name string           `json:"name"`
		workerConfigItem
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Name == "" {
		http.Error(w, "name is required", http.StatusBadRequest)
		return
	}
	if _, exists := h.store.Worker(body.Name); exists {
		http.Error(w, "worker already exists", http.StatusConflict)
		return
	}
	wc := config.WorkerConfig{
		Image:   body.Image,
		Cmd:     body.Cmd,
		Labels:  body.Labels,
		Secrets: body.Secrets,
	}
	if err := h.store.SetWorker(body.Name, wc); err != nil {
		log.Error().Err(err).Str("worker", body.Name).Msg("failed to save worker")
		http.Error(w, "failed to save worker", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *WorkerDefinitionsHandler) Update(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	var body workerConfigItem
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if _, exists := h.store.Worker(name); !exists {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}
	wc := config.WorkerConfig{
		Image:   body.Image,
		Cmd:     body.Cmd,
		Labels:  body.Labels,
		Secrets: body.Secrets,
	}
	if err := h.store.SetWorker(name, wc); err != nil {
		log.Error().Err(err).Str("worker", name).Msg("failed to save worker")
		http.Error(w, "failed to save worker", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *WorkerDefinitionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")
	if err := h.store.DeleteWorker(name); err != nil {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func toWorkerConfigItem(wc config.WorkerConfig) workerConfigItem {
	secrets := wc.Secrets
	if secrets == nil {
		secrets = []string{}
	}
	return workerConfigItem{
		Image:   wc.Image,
		Cmd:     wc.Cmd,
		Labels:  wc.Labels,
		Secrets: secrets,
	}
}
