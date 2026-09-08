package persistence_test

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/joho/godotenv"

	"github.com/javiermm8/open-gym-vault/internal/db"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

// newTestStore opens a real connection to the database configured by
// DATABASE_URL, skipping the test if it's not set (so `go test ./...`
// doesn't fail in environments with no database available).
func newTestStore(t *testing.T) *persistence.Store {
	t.Helper()
	_ = godotenv.Load("../../.env") // best-effort; ignored if missing

	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		t.Skip("DATABASE_URL not set, skipping integration test")
	}

	store, err := persistence.NewStore(context.Background(), dbURL)
	if err != nil {
		t.Fatalf("failed to connect to database: %v", err)
	}
	t.Cleanup(store.Close)
	return store
}

// setupTestUserAndExercise creates a throwaway user and a custom exercise
// owned by that user. Deleting the user cascades to the exercise, the
// session, and its activities — so cleanup is just one delete.
func setupTestUserAndExercise(t *testing.T, ctx context.Context, store *persistence.Store) (uuid.UUID, uuid.UUID) {
	t.Helper()

	user, err := store.Queries.CreateUser(ctx, db.CreateUserParams{
		Username:     "test_" + uuid.NewString(),
		DisplayName:  "Test User",
		PasswordHash: "not-a-real-hash",
	})
	if err != nil {
		t.Fatalf("failed to create test user: %v", err)
	}
	userID, err := persistence.FromPgUUID(user.ID)
	if err != nil {
		t.Fatalf("failed to convert user ID: %v", err)
	}

	exercise, err := store.Queries.CreateCustomExercise(ctx, db.CreateCustomExerciseParams{
		UserID:           user.ID,
		Name:             "Bench Press",
		AlternativeNames: []string{"Barbell Bench Press"},
	})
	if err != nil {
		t.Fatalf("failed to create test exercise: %v", err)
	}
	exerciseID, err := persistence.FromPgUUID(exercise.ID)
	if err != nil {
		t.Fatalf("failed to convert exercise ID: %v", err)
	}

	t.Cleanup(func() {
		// Deleting the user cascades to their custom exercise, and to any
		// sessions/activities created against that user in the test.
		if err := store.DeleteUserCascade(ctx, userID); err != nil {
			t.Errorf("cleanup: failed to delete test user %s: %v", userID, err)
		}
	})

	return userID, exerciseID
}

func TestCreateSessionWithActivities_Valid(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	userID, exerciseID := setupTestUserAndExercise(t, ctx, store)

	now := time.Now()
	reps := int32(8)
	weight := float32(60)

	session, activities, err := store.CreateSessionWithActivities(ctx, persistence.NewSession{
		UserID:      userID,
		SessionType: "strength",
		StartTime:   now,
		EndTime:     now.Add(45 * time.Minute),
		Activities: []persistence.NewActivity{
			{
				ExerciseID:   &exerciseID,
				ActivityType: "exercise",
				Reps:         &reps,
				Weight:       &weight,
				StartTime:    now,
				EndTime:      now.Add(30 * time.Second),
			},
			{
				ActivityType: "rest",
				StartTime:    now.Add(30 * time.Second),
				EndTime:      now.Add(90 * time.Second),
			},
		},
	})

	if err != nil {
		t.Fatalf("expected valid session to succeed, got error: %v", err)
	}
	if len(activities) != 2 {
		t.Fatalf("expected 2 activities to be created, got %d", len(activities))
	}

	// Confirm it actually landed in the database, not just returned in memory.
	stored, err := store.Queries.ListActivitiesBySession(ctx, session.ID)
	if err != nil {
		t.Fatalf("failed to list activities: %v", err)
	}
	if len(stored) != 2 {
		t.Fatalf("expected 2 activities persisted, found %d", len(stored))
	}
}

func TestCreateSessionWithActivities_Invalid_RollsBackCompletely(t *testing.T) {
	ctx := context.Background()
	store := newTestStore(t)
	userID, _ := setupTestUserAndExercise(t, ctx, store)

	sessionsBefore, err := store.Queries.ListSessionsByUser(ctx, persistence.ToPgUUID(userID))
	if err != nil {
		t.Fatalf("failed to list sessions before test: %v", err)
	}

	now := time.Now()

	// This activity claims to be type "exercise" but has no ExerciseID —
	// violates the exercise_id_matches_type CHECK constraint.
	_, _, err = store.CreateSessionWithActivities(ctx, persistence.NewSession{
		UserID:      userID,
		SessionType: "strength",
		StartTime:   now,
		EndTime:     now.Add(45 * time.Minute),
		Activities: []persistence.NewActivity{
			{
				ActivityType: "rest",
				StartTime:    now,
				EndTime:      now.Add(60 * time.Second),
			},
			{
				ActivityType: "exercise", // invalid: no ExerciseID set
				StartTime:    now.Add(60 * time.Second),
				EndTime:      now.Add(90 * time.Second),
			},
		},
	})

	if err == nil {
		t.Fatal("expected an error due to constraint violation, got nil")
	}

	// The real test: confirm NOTHING was persisted, including the first
	// (individually valid) rest activity and the session itself.
	sessionsAfter, err := store.Queries.ListSessionsByUser(ctx, persistence.ToPgUUID(userID))
	if err != nil {
		t.Fatalf("failed to list sessions after test: %v", err)
	}
	if len(sessionsAfter) != len(sessionsBefore) {
		t.Fatalf("expected rollback to leave session count unchanged: before=%d after=%d",
			len(sessionsBefore), len(sessionsAfter))
	}
}
