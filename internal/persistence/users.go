package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/javiermm8/open-gym-vault/internal/db"
)

// DeleteUserCascade deletes a user and everything that depends on them:
// their activities, their custom exercises, their sessions, and finally
// the user row itself.
//
// This can't be left to a single `DELETE FROM users` relying purely on
// ON DELETE CASCADE, because activities.exercise_id is intentionally
// ON DELETE RESTRICT (to protect *global* exercises from being deleted
// while other users still reference them). That RESTRICT can conflict
// with the CASCADE path from users -> exercises if Postgres processes it
// before the users -> sessions -> activities cascade has cleared the
// references — so we control the order explicitly instead.
func (s *Store) DeleteUserCascade(ctx context.Context, userID uuid.UUID) error {
	pgID := ToPgUUID(userID)

	return s.WithTx(ctx, func(q *db.Queries) error {
		if err := q.DeleteActivitiesByUser(ctx, pgID); err != nil {
			return fmt.Errorf("deleting user's activities: %w", err)
		}
		if err := q.DeleteExercisesByUser(ctx, pgID); err != nil {
			return fmt.Errorf("deleting user's custom exercises: %w", err)
		}
		if err := q.DeleteUser(ctx, pgID); err != nil {
			return fmt.Errorf("deleting user: %w", err)
		}
		return nil
	})
}
