package main

import (
	"maps"
	"net/http"
	"os"
	"slices"

	"agentsmanager/pkg/config"
	"agentsmanager/pkg/handler"

	"github.com/algorath-software/workerd/pkg/client"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func main() {
	log.Logger = log.Output(zerolog.ConsoleWriter{Out: os.Stderr})

	cfg, err := config.Load("etc/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	workerdClient, err := client.New(cfg.Workerd.Addr)
	if err != nil {
		log.Fatal().Err(err).Str("addr", cfg.Workerd.Addr).Msg("failed to connect to workerd")
	}
	defer workerdClient.Close()

	workerNames := slices.Sorted(maps.Keys(cfg.Workers))

	mux := http.NewServeMux()
	mux.Handle("GET /workers", handler.NewWorkersHandler(workerNames))
	mux.Handle("POST /deploy", handler.NewDeployHandler(workerdClient, cfg.Workers, cfg.Secrets))
	mux.Handle("GET /containers", handler.NewContainersHandler(workerdClient))
	mux.Handle("GET /containers/{id}", handler.NewContainerHandler(workerdClient))
	mux.Handle("GET /containers/{id}/logs", handler.NewLogsHandler(workerdClient))
	mux.Handle("DELETE /containers/{id}", handler.NewRemoveHandler(workerdClient))

	log.Info().Str("addr", cfg.Server.Addr).Msg("server listening")
	if err := http.ListenAndServe(cfg.Server.Addr, mux); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
