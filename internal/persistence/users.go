package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/javiermm8/open-gym-vault/internal/db"
)

type NewUser struct {
	Username      string
	DisplayName   string
	PasswordHash  string
	Bio           *string
	Sex           *string
	Birthday      time.Time
	CreatedAt     time.Time
	LastUpdatedAt time.Time
}

func (s *Store) CreateUser(ctx context.Context, in NewUser) (db.User, error) {
	var user db.User

	err := s.WithTx(ctx, func(q *db.Queries) error {
		var err error
		user, err = q.CreateUser(ctx, db.CreateUserParams{
			Username:      in.Username,
			DisplayName:   in.DisplayName,
			PasswordHash:  in.PasswordHash,
			Bio:           ToPgTextPtr(in.Bio),
			Sex:           ToPgTextPtr(in.Sex),
			Birthday:      ToPgDate(in.Birthday),
			CreatedAt:     ToPgTimestamptz(time.Now()),
			LastUpdatedAt: ToPgTimestamptz(time.Time{}),
		})
		if err != nil {
			return fmt.Errorf("creating user: %w", err)
		}

		return nil
	})
	if err != nil {
		return db.User{}, err
	}

	return user, nil
}

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
