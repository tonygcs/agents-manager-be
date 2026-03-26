package handler

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/google/uuid"
)

type deployRequest struct {
	WorkerName string `json:"workerName"`
}

type deployResponse struct {
	ContainerID string `json:"containerId"`
}

// workerOptions maps a worker name to a function that builds DeployOptions for a given container name and env.
var workerOptions = map[string]func(name string, env []string) client.DeployOptions{
	"opencode-node": func(name string, env []string) client.DeployOptions {
		return client.DeployOptions{
			Image: "opencode-worker-node:latest",
			Name:  name,
			Cmd:   []string{"opencode", "web", "--hostname", "0.0.0.0", "--mdns-domain", fmt.Sprintf("%s.localhost", name)},
			Env:   env,
			Labels: map[string]string{
				"traefik.enable":                                              "true",
				"traefik.http.routers." + name + ".rule":                      "Host(`" + name + ".localhost`)",
				"traefik.http.services." + name + ".loadbalancer.server.port": "4096",
			},
		}
	},
}

type DeployHandler struct {
	client      *client.Client
	githubToken string
}

func NewDeployHandler(c *client.Client, githubToken string) *DeployHandler {
	return &DeployHandler{client: c, githubToken: githubToken}
}

func (h *DeployHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req deployRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid JSON", http.StatusBadRequest)
		return
	}

	buildOptions, ok := workerOptions[req.WorkerName]
	if !ok {
		http.Error(w, "worker not found", http.StatusNotFound)
		return
	}

	name := uuid.NewString()
	env := []string{
		"GITHUB_TOKEN=" + h.githubToken,
	}
	result, err := h.client.Deploy(r.Context(), buildOptions(name, env))
	if err != nil {
		log.Printf("deploy failed for worker %s: %v", req.WorkerName, err)
		http.Error(w, "deploy failed", http.StatusInternalServerError)
		return
	}

	log.Printf("deployed worker %s: container %s", req.WorkerName, result.ID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(deployResponse{ContainerID: result.ID})
}
