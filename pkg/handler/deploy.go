package handler

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"agentsmanager/pkg/config"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/google/uuid"
)

type deployRequest struct {
	WorkerName string `json:"workerName"`
}

type deployResponse struct {
	ContainerID string `json:"containerId"`
}

func applyTemplate(s, name string) string {
	return strings.ReplaceAll(s, "{{name}}", name)
}

type DeployHandler struct {
	client  *client.Client
	workers map[string]config.WorkerConfig
	secrets map[string]string
}

func NewDeployHandler(c *client.Client, workers map[string]config.WorkerConfig, secrets map[string]string) *DeployHandler {
	return &DeployHandler{client: c, workers: workers, secrets: secrets}
}

func (h *DeployHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	workerCfg, ok := h.workers[req.WorkerName]
	if !ok {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}

	name := uuid.NewString()

	cmd := make([]string, len(workerCfg.Cmd))
	for i, part := range workerCfg.Cmd {
		cmd[i] = applyTemplate(part, name)
	}

	labels := make(map[string]string, len(workerCfg.Labels))
	for k, v := range workerCfg.Labels {
		labels[applyTemplate(k, name)] = applyTemplate(v, name)
	}

	env := make([]string, 0, len(workerCfg.Secrets))
	for _, key := range workerCfg.Secrets {
		env = append(env, key+"="+h.secrets[key])
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
		log.Printf("deploy failed for worker %s: %v", req.WorkerName, err)
		http.Error(w, "deploy failed", http.StatusInternalServerError)
		return
	}

	log.Printf("deployed worker %s: container %s", req.WorkerName, result.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployResponse{ContainerID: result.ID})
}
