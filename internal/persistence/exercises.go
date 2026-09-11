package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/db"
)

type NewExercise struct {
	UserID           string
	Name             string
	AlternativeNames *[]string
	Explanation      *string
	ClientS          *json.RawMessage
	CreatedAt        time.Time
	LastUpdatedAt    time.Time
}

func (s *Store) CreateExercise(ctx context.Context, in NewExercise) (db.Exercise, error) {
	var exercise db.Exercise

	var altNames []string
	if in.AlternativeNames != nil {
		altNames = *in.AlternativeNames
	}

	var clientStuff json.RawMessage
	if in.ClientS != nil {
		clientStuff = *in.ClientS
	}

	err := s.WithTx(ctx, func(q *db.Queries) error {
		var err error
		exercise, err = q.CreateCustomExercise(ctx, db.CreateCustomExerciseParams{
			UserID:           ToPgText(string(in.UserID)),
			Name:             in.Name,
			AlternativeNames: altNames,
			Explanation:      ToPgTextPtr(in.Explanation),
			ClientS:          clientStuff,
		})
		if err != nil {
			return fmt.Errorf("creating exercise: %w", err)
		}
		return nil
	})
	if err != nil {
		return db.Exercise{}, err
	}

	return exercise, nil
}

func (s *Store) QueryExecise(ctx context.Context, exerciseID string) (db.Exercise, error) {
	exercise, err := s.Queries.GetExerciseByID(ctx, exerciseID)
	if err != nil {
		return db.Exercise{}, fmt.Errorf("Quering exercise: %w", err)
	}

	return exercise, nil
}

func (s *Store) QueryExercises(ctx context.Context, userID string) ([]db.Exercise, error) {
	exercises, err := s.Queries.ListExercisesForUser(ctx, ToPgText(userID))
	if err != nil {
		return []db.Exercise{}, fmt.Errorf("Quering exercises: %w", err)

	}

	return exercises, nil
}
