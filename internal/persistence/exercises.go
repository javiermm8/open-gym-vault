package persistence

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/javiermm8/open-gym-vault/internal/db"
)

type NewExercise struct {
	UserID           *uuid.UUID
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

	err := s.WithTx(ctx, func(q *db.Queries) error {
		var err error
		exercise, err = q.CreateCustomExercise(ctx, db.CreateCustomExerciseParams{
			UserID:           ToPgUUIDPtr(in.UserID),
			Name:             in.Name,
			AlternativeNames: altNames,
			Explanation:      ToPgTextPtr(in.Explanation),
			ClientS:          *in.ClientS,
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
