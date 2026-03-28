package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"text/template"

	"agentsmanager/pkg/config"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type deployRequest struct {
	WorkerName string `json:"workerName"`
}

type deployResponse struct {
	ContainerID string `json:"containerId"`
}

var templateFuncs = template.FuncMap{
	"split": strings.Split,
}

func applyTemplate(s, name string) (string, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"name": name}); err != nil {
		return "", err
	}
	return buf.String(), nil
}

type DeployHandler struct {
	client *client.Client
	store  *config.Store
}

func NewDeployHandler(c *client.Client, store *config.Store) *DeployHandler {
	return &DeployHandler{client: c, store: store}
}

func (h *DeployHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	workerCfg, ok := h.store.Worker(req.WorkerName)
	if !ok {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}

	secrets := h.store.Secrets()
	name := uuid.NewString()

	cmd := make([]string, len(workerCfg.Cmd))
	for i, part := range workerCfg.Cmd {
		var err error
		cmd[i], err = applyTemplate(part, name)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in cmd")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
	}

	labels := make(map[string]string, len(workerCfg.Labels))
	for k, v := range workerCfg.Labels {
		key, err := applyTemplate(k, name)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in label key")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
		val, err := applyTemplate(v, name)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in label value")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
		labels[key] = val
	}

	env := make([]string, 0, len(workerCfg.Secrets))
	for _, key := range workerCfg.Secrets {
		env = append(env, key+"="+secrets[key])
	}

	opts := client.DeployOptions{
		Image:  workerCfg.Image,
		Name:   name,
		Cmd:    cmd,
		Env:    env,
		Labels: labels,
	}

	result, err := h.client.Deploy(r.Context(), opts)
	if err != nil {
		log.Error().Err(err).Str("worker", req.WorkerName).Msg("deploy failed")
		http.Error(w, "deploy failed", http.StatusInternalServerError)
		return
	}

	log.Info().Str("worker", req.WorkerName).Str("container", result.ID).Msg("deployed")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployResponse{ContainerID: result.ID})
}
