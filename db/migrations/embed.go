// internal/db/migrations/embed.go
//
// Package migrations embeds the SQL migration files so they can be shipped
// inside the compiled binary rather than read from disk at runtime.

package migrations

import "embed"

//go:embed *.sql
var FS embed.FS
