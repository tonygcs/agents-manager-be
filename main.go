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

	store, err := config.NewStore("etc/config.yaml")
	if err != nil {
		log.Fatal().Err(err).Msg("failed to load config")
	}

	workerdClient, err := client.New(store.WorkerdAddr())
	if err != nil {
		log.Fatal().Err(err).Str("addr", store.WorkerdAddr()).Msg("failed to connect to workerd")
	}
	defer workerdClient.Close()

	workerNames := slices.Sorted(maps.Keys(store.Workers()))

	mux := http.NewServeMux()
	secrets := handler.NewSecretsHandler(store)
	mux.HandleFunc("GET /config/secrets", secrets.List)
	mux.HandleFunc("POST /config/secrets", secrets.Create)
	mux.HandleFunc("GET /config/secrets/{key}", secrets.Get)
	mux.HandleFunc("PUT /config/secrets/{key}", secrets.Update)
	mux.HandleFunc("DELETE /config/secrets/{key}", secrets.Delete)
	workerDefs := handler.NewWorkerDefinitionsHandler(store)
	mux.HandleFunc("GET /config/workers", workerDefs.List)
	mux.HandleFunc("POST /config/workers", workerDefs.Create)
	mux.HandleFunc("GET /config/workers/{name}", workerDefs.Get)
	mux.HandleFunc("PUT /config/workers/{name}", workerDefs.Update)
	mux.HandleFunc("DELETE /config/workers/{name}", workerDefs.Delete)
	mux.Handle("GET /workers", handler.NewWorkersHandler(workerNames))
	mux.Handle("POST /deploy", handler.NewDeployHandler(workerdClient, store.Workers(), store.Secrets()))
	mux.Handle("GET /containers", handler.NewContainersHandler(workerdClient))
	mux.Handle("GET /containers/{id}", handler.NewContainerHandler(workerdClient))
	mux.Handle("GET /containers/{id}/logs", handler.NewLogsHandler(workerdClient))
	mux.Handle("DELETE /containers/{id}", handler.NewRemoveHandler(workerdClient))

	log.Info().Str("addr", store.ServerAddr()).Msg("server listening")
	if err := http.ListenAndServe(store.ServerAddr(), mux); err != nil {
		log.Fatal().Err(err).Msg("server stopped")
	}
}
