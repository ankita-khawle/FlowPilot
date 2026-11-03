package main

import (
	"log"
	"net/http"
	"time"

	"flowpilot/backend/internal/config"
	"flowpilot/backend/internal/httpapi"
)

func main() {
	cfg := config.Load()
	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           httpapi.NewRouter(),
		ReadHeaderTimeout: 5 * time.Second,
	}

	log.Printf("FlowPilot API listening on %s", cfg.HTTPAddr)
	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("start API server: %v", err)
	}
}
