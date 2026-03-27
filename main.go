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

	http.Handle("/deploy", handler.NewDeployHandler(workerdClient, cfg.GitHub.Token))
	http.Handle("/containers", handler.NewContainersHandler(workerdClient))
	http.Handle("/containers/", handler.NewLogsHandler(workerdClient))

	log.Printf("server listening on %s", cfg.Server.Addr)
	log.Fatal(http.ListenAndServe(cfg.Server.Addr, nil))
}
