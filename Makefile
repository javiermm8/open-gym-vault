.PHONY: db-up db-down migrate-up migrate-down sqlc run

# Start Postgres in the background
db-up:
	docker compose up -d

# Stop Postgres (data persists in the named volume)
db-down:
	docker compose down

# Apply all pending migrations
migrate-up:
	go run ./cmd/migrate up

# Roll back the most recent migration
migrate-down:
	go run ./cmd/migrate down

# Regenerate Go code from db/migrations + db/queries
sqlc:
	sqlc generate

# Run the app
run:
	go run ./cmd/gymvault
