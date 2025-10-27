package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"redis-from-scratch/internal/server"
	"redis-from-scratch/pkg/config"
)

func main() {
	configPath := flag.String("config", "", "path to config file")
	port := flag.Int("port", 6379, "port to listen on")
	flag.Parse()

	cfg := config.DefaultConfig()
	if *configPath != "" {
		loadedCfg, err := config.LoadFromFile(*configPath)
		if err != nil {
			log.Printf("Failed to load config: %v, using defaults", err)
		} else {
			cfg = loadedCfg
		}
	}
	if *port != 6379 {
		cfg.Port = *port
	}

	srv := server.New(cfg)

	// Handle graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		log.Println("Shutting down server...")
		srv.Stop()
		os.Exit(0)
	}()

	log.Printf("Starting Redis server on port %d", cfg.Port)
	if err := srv.Start(); err != nil {
		log.Fatal(err)
	}
}
