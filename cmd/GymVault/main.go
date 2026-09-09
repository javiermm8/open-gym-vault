// src/GymVault/main.go
package main

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"github.com/javiermm8/open-gym-vault/internal/api"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func main() {
	// Load .env if exists. Does nothing if it doesn't(env variables could be loaded via docker or smth else)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: found .env but failed to load it: %v", err)
	}

	// Get postgress url from env variables
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Get listening url from env variables
	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	// Some sort of black magic
	ctx := context.Background()

	// Open a connection pool to postgress and pings it
	store, err := persistence.NewStore(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	// api.Server, holds the dependencies handler needs
	server := api.New(store)

	httpServer := api.NewHTTPServer(addr, server)

	// On a gorutine so it we can still listen to sigterms and other stop signals
	go func() {
		log.Printf("listening on %s", addr)
		if err := httpServer.Start(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("server error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	<-stop

	log.Println("shutting down...")
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := httpServer.Shutdown(shutdownCtx); err != nil {
		log.Printf("error during shutdown: %v", err)
	}
	return
}
