// src/internal/api/types.go
package api

import (
	"database/sql"
	"encoding/json"
	"time"
)

// All units are in the metric system
// ClientS stands for custom info such as client configs and extra features added by the client supposed to be stored as a json blob.

type Exercise struct {
	ID               string
	UserID           sql.Null[string]
	Name             string
	AlternativeNames sql.Null[[]string]
	Explanation      sql.Null[string]
	// MuscleSplit      string

	CreatedAt     time.Time
	LastUpdatedAt time.Time

	ClientS sql.Null[json.RawMessage]
}

// type MuscleSplit struct {

// }

type User struct {
	ID           string
	Username     string
	DisplayName  string
	PasswordHash string
	Bio          sql.Null[string]
	Sex          sql.Null[string]
	Birthday     sql.Null[time.Time]
	// TODO: Templates(probably some type of session)
	CreatedAt     time.Time
	LastUpdatedAt time.Time

	ClientS sql.Null[json.RawMessage]
}

type Session struct {
	ID          string
	UserID      string
	SessionType string
	StartTime   time.Time
	EndTime     time.Time
	TotalTime   time.Duration // Calculated in the backend at time of submition.
	TotalWeight int           // Calculated in the backend at time of submition.
	// TODO: Muscle split. Calculated in the backend at time of submition.
	OverallPerceivedEffort sql.Null[int]
	// TODO: bpms
	BurnedCals sql.Null[int]
	// TODO: Muscle split (probably its own type)
	UserNotes sql.Null[string]

	CreatedAt     time.Time
	LastUpdatedAt time.Time

	ClientS sql.Null[json.RawMessage]
}

type Activity struct {
	ID              string
	SessionID       string
	ExcerciseID     sql.Null[string]
	ActivityType    string // Excersise or rest
	Reps            sql.Null[int]
	Weight          sql.Null[float32] // in kg
	SortOrder       int               // would name it index but SQL chose that name first(not reserved but might bue confusing)
	StartTime       time.Time
	EndTime         time.Time
	TotalTime       time.Duration // Calculated in the backend at time of submition.
	PerceivedEffort sql.Null[int]

	CreatedAt     time.Time
	LastUpdatedAt time.Time

	ClientS sql.Null[json.RawMessage]
}
