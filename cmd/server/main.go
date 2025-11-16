package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"AvitoTest/internal/server"
)

func main() {
	envPath := os.Getenv("ENV_FILE")
	if envPath == "" {
		envPath = "../../environment/.env"
	}

	srv, err := server.NewServer(envPath)
	if err != nil {
		log.Fatalf("Failed to create server: %v", err)
	}

	go func() {
		if err := srv.Start(); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}
