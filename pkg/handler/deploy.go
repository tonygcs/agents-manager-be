package handler

import (
	"bytes"
	"encoding/json"
	"net"
	"net/http"
	"strings"
	"text/template"

	"agentsmanager/pkg/config"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
)

type deployRequest struct {
	WorkerName string            `json:"workerName"`
	Env        map[string]string `json:"env"`
}

type deployResponse struct {
	ContainerID string `json:"containerId"`
}

var templateFuncs = template.FuncMap{
	"split": strings.Split,
	"rootdomain": func(host string) string {
		knownTLDs := map[string]bool{
			"com": true, "net": true, "org": true, "edu": true,
			"gov": true, "io": true, "es": true, "uk": true,
			"de": true, "fr": true, "it": true, "co": true,
			"us": true, "ai": true, "app": true, "dev": true,
			"me": true, "tv": true, "info": true, "biz": true,
		}
		parts := strings.Split(host, ".")
		if len(parts) == 1 {
			return host
		}
		if knownTLDs[parts[len(parts)-1]] {
			return strings.Join(parts[len(parts)-2:], ".")
		}
		return parts[len(parts)-1]
	},
}

func applyTemplate(s, name, domain string) (string, error) {
	tmpl, err := template.New("").Funcs(templateFuncs).Parse(s)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, map[string]string{"name": name, "domain": domain}); err != nil {
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
	domain, _, err := net.SplitHostPort(r.Host)
	if err != nil {
		domain = r.Host
	}

	cmd := make([]string, len(workerCfg.Cmd))
	for i, part := range workerCfg.Cmd {
		var err error
		cmd[i], err = applyTemplate(part, name, domain)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in cmd")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
	}

	labels := make(map[string]string, len(workerCfg.Labels))
	for k, v := range workerCfg.Labels {
		key, err := applyTemplate(k, name, domain)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in label key")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
		val, err := applyTemplate(v, name, domain)
		if err != nil {
			log.Error().Err(err).Str("worker", req.WorkerName).Msg("template error in label value")
			http.Error(w, "deploy failed", http.StatusInternalServerError)
			return
		}
		labels[key] = val
	}

	envMap := make(map[string]string, len(workerCfg.Secrets)+len(req.Env))
	for _, key := range workerCfg.Secrets {
		envMap[key] = secrets[key]
	}
	for k, v := range req.Env {
		envMap[k] = v
	}
	env := make([]string, 0, len(envMap))
	for k, v := range envMap {
		env = append(env, k+"="+v)
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
