package persistence

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/javiermm8/open-gym-vault/internal/db"
)

func (s *Store) CalculatePR(ctx context.Context, userID, exerciseID string) (int, error) {
	pr, err := s.Queries.GetMaxWeightByExerciseID(ctx, db.GetMaxWeightByExerciseIDParams{
		UserIDInAct: userID,
		ExerciseID:  ToPgText(exerciseID),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return 0, nil
		}
		return -1, err
	}
	return int(pr), nil
}
