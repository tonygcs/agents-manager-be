package main

import (
	"log"
	"net/http"

	"agentsmanager/pkg/config"
	"agentsmanager/pkg/handler"

	"github.com/algorath-software/workerd/pkg/client"
)

func main() {
	cfg, err := config.Load("etc/config.yaml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	workerdClient, err := client.New(cfg.Workerd.Addr)
	if err != nil {
		log.Fatalf("failed to connect to workerd at %s: %v", cfg.Workerd.Addr, err)
	}
	defer workerdClient.Close()

	mux := http.NewServeMux()
	mux.Handle("GET /workers", handler.NewWorkersHandler())
	mux.Handle("POST /deploy", handler.NewDeployHandler(workerdClient, cfg.GitHub.Token))
	mux.Handle("GET /containers", handler.NewContainersHandler(workerdClient))
	mux.Handle("GET /containers/{id}/logs", handler.NewLogsHandler(workerdClient))
	mux.Handle("DELETE /containers/{id}", handler.NewRemoveHandler(workerdClient))

	log.Printf("server listening on %s", cfg.Server.Addr)
	log.Fatal(http.ListenAndServe(cfg.Server.Addr, mux))
}
