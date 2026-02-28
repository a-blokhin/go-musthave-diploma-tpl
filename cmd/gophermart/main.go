package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"go-musthave-diploma-tpl/internal/config"
	"go-musthave-diploma-tpl/internal/di/app"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.ParseConfig()

	di := app.DI{}
	if err := di.Init(cfg); err != nil {
		log.Fatalf("Failed to initialize application: %v", err)
	}

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	errCh := make(chan error, 1)

	go func() {
		log.Printf("Starting server on %s", cfg.ServerAddress)
		if err := di.StartServer(ctx); err != nil {
			errCh <- fmt.Errorf("server error: %w", err)
		}
	}()

	select {
	case <-stop:
		log.Println("Shutting down server gracefully...")
	case err := <-errCh:
		log.Printf("Error: %v", err)
	}

	shutdownCtx, shutdownCancel := context.WithTimeout(ctx, 10*time.Second)
	defer shutdownCancel()

	if err := di.StopServer(shutdownCtx); err != nil {
		log.Fatalf("Server shutdown failed: %v", err)
	}

	log.Println("Server stopped successfully")
}
