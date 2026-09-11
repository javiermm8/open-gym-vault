package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"

	"github.com/javiermm8/open-gym-vault/internal/auth"
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
	ClientS       *json.RawMessage
}

var ErrInvalidCredentials = errors.New("invalid username or password")
var ErrTokenInvalidOrExpired = errors.New("token invalid or expired")

func (s *Store) Register(ctx context.Context, username, displayName, password string) (db.User, error) {
	hash, err := auth.HashPassword(password)
	if err != nil {
		return db.User{}, err
	}

	user, err := s.Queries.CreateUser(ctx, db.CreateUserParams{
		Username:     username,
		DisplayName:  displayName,
		PasswordHash: hash,
		CreatedAt:    ToPgTimestamptz(time.Now()),
	})
	if err != nil {
		return db.User{}, fmt.Errorf("creating usr: %w", err)
	}
	return user, nil
}

func (s *Store) Login(ctx context.Context, username, password string) (rawToken string, expiresAt time.Time, user db.User, err error) {
	user, err = s.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		return "", time.Time{}, db.User{}, ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return "", time.Time{}, db.User{}, ErrInvalidCredentials
	}

	raw, hash, err := auth.GenrateRawToken()
	if err != nil {
		return "", time.Time{}, db.User{}, fmt.Errorf("generating raw token: %w", err)
	}
	expiresAt = time.Now().Add(auth.TokenTTL)

	if _, err := s.Queries.CreateAuthToken(ctx, db.CreateAuthTokenParams{
		UserID:    user.ID,
		TokenHash: hash,
		ExpiresAt: ToPgTimestamptz(expiresAt),
	}); err != nil {
		return "", time.Time{}, db.User{}, fmt.Errorf("creating auth token: %w", err)
	}

	return raw, expiresAt, user, nil
}

func (s *Store) ExtendTokenExpiry(ctx context.Context, rawToken string) (uuid.UUID, error) {
	hash := auth.HashToken(rawToken)

	token, err := s.Queries.GetAuthTokenByHash(ctx, hash)
	if err != nil {
		return uuid.UUID{}, ErrTokenInvalidOrExpired
	}

	if token.ExpiresAt.Time.Before(time.Now()) {
		if err = s.Queries.DeleteAuthToken(ctx, auth.HashToken(hash)); err != nil {
			log.Printf("deleting auth token: %w", err)
		}
		return uuid.UUID{}, ErrTokenInvalidOrExpired
	}

	newExpiry := time.Now().Add(auth.TokenTTL)
	if err := s.Queries.RefreshAuthTokenExpiry(ctx, db.RefreshAuthTokenExpiryParams{
		ID:        token.ID,
		ExpiresAt: ToPgTimestamptz(newExpiry),
	}); err != nil {
		return uuid.UUID{}, fmt.Errorf("refreshing auth token: %w", err)
	}

	return FromPgUUID(token.UserID)
}

func (s *Store) Logout(ctx context.Context, rawToken string) error {
	return s.Queries.DeleteAuthToken(ctx, auth.HashToken(rawToken))
}

func (s *Store) CreateUser(ctx context.Context, in NewUser) (db.User, error) {
	var user db.User

	var clientStuff json.RawMessage
	if in.ClientS != nil {
		clientStuff = *in.ClientS
	}

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
			ClientS:       clientStuff,
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

func (s *Store) QueryUser(ctx context.Context, id string) (db.User, error) {
	userUUID, err := uuid.Parse(id)
	if err != nil {
		return db.User{}, fmt.Errorf("Quering user: Parse uuid: %w", err)
	}
	userID := ToPgUUID(userUUID)

	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return db.User{}, fmt.Errorf("Quering user: %w", err)
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
