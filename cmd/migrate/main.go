// Command migrate applies or rolls back database migrations for OpenGymVault.
//
// Usage:
//
//	go run ./cmd/migrate up            # apply all pending migrations
//	go run ./cmd/migrate down          # roll back the most recent migration
//	go run ./cmd/migrate down-all      # roll back all migrations
//	go run ./cmd/migrate version       # print current migration version
//	go run ./cmd/migrate force <ver>   # force migration version (use -1 to clear dirty state)
//
// Reads the connection string from the DATABASE_URL environment variable.
package main

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/joho/godotenv"

	"github.com/javiermm8/open-gym-vault/db/migrations"
)

func main() {
	if len(os.Args) < 2 {
		log.Fatal("usage: migrate <up|down|down-all|version>")
	}

	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("warning: found .env but failed to load it: %v", err)
	}

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	src, err := iofs.New(migrations.FS, ".")
	if err != nil {
		log.Fatalf("failed to load embedded migrations: %v", err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dbURL)
	if err != nil {
		log.Fatalf("failed to initialize migrator: %v", err)
	}
	defer m.Close()

	switch os.Args[1] {
	case "up":
		err = m.Up()
	case "down":
		err = m.Steps(-1)
	case "down-all":
		err = m.Down()
	case "version":
		version, dirty, verr := m.Version()
		if verr != nil {
			log.Fatalf("failed to read version: %v", verr)
		}
		fmt.Printf("version=%d dirty=%v\n", version, dirty)
		return
	case "force":
		if len(os.Args) < 3 {
			log.Fatal("usage: migrate force <version>")
		}
		var version int
		if _, err := fmt.Sscanf(os.Args[2], "%d", &version); err != nil {
			log.Fatalf("invalid version %q: %v", os.Args[2], err)
		}
		err = m.Force(version)
	default:
		log.Fatalf("unknown command %q", os.Args[1])
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("migration failed: %v", err)
	}
	fmt.Println("migrations applied successfully")
}
