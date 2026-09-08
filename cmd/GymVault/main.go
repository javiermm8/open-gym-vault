// src/GymVault/main.go
package main

import (
	"context"
	"log"
	"os"

	"github.com/joho/godotenv"

	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func main() {
	// load config from config file
	//
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: found .env but failed to load it: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	ctx := context.Background()

	store, err := persistence.NewStore(ctx, dbURL)
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	defer store.Close()

	log.Println("connected to database, GymVault starting...")

	// TODO: wire up internal/api routes here, using store.Queries
	// to call the generated methods (store.Queries.CreateUser(...), etc).

	return
}
