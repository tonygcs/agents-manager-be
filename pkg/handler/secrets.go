package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"agentsmanager/pkg/config"

	"github.com/rs/zerolog/log"
)

type SecretsHandler struct {
	store *config.Store
}

func NewSecretsHandler(store *config.Store) *SecretsHandler {
	return &SecretsHandler{store: store}
}

func (h *SecretsHandler) List(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Clear all secret values
	s := h.store.Secrets()
	for k := range s {
		s[k] = ""
	}

	json.NewEncoder(w).Encode(s)
}

func (h *SecretsHandler) Create(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if body.Key == "" {
		http.Error(w, "key is required", http.StatusBadRequest)
		return
	}
	if _, exists := h.store.Secret(body.Key); exists {
		http.Error(w, "secret already exists", http.StatusConflict)
		return
	}
	if err := h.store.SetSecret(body.Key, body.Value); err != nil {
		log.Error().Err(err).Str("key", body.Key).Msg("failed to save secret")
		http.Error(w, "failed to save secret", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusCreated)
}

func (h *SecretsHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	var body struct {
		Value string `json:"value"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}
	if _, exists := h.store.Secret(key); !exists {
		http.Error(w, "secret not found", http.StatusNotFound)
		return
	}
	if err := h.store.SetSecret(key, body.Value); err != nil {
		log.Error().Err(err).Str("key", key).Msg("failed to save secret")
		http.Error(w, "failed to save secret", http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *SecretsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := r.PathValue("key")
	err := h.store.DeleteSecret(key)
	if err == nil {
		w.WriteHeader(http.StatusNoContent)
		return
	}
	var inUse *config.ErrInUse
	if errors.As(err, &inUse) {
		http.Error(w, err.Error(), http.StatusConflict)
		return
	}
	http.Error(w, "secret not found", http.StatusNotFound)
}
