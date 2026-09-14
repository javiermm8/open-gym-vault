package persistence

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/jackc/pgx/v5"
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

type UpdatedUser struct {
	Username    string
	DisplayName string
	Bio         *string
	Sex         *string
	Birthday    time.Time
	ClientS     *json.RawMessage
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

func (s *Store) ChangePassword(ctx context.Context, userID, password, newPassword string) (bool, error) {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return false, err
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return true, nil
	}

	hashNew, err := auth.HashPassword(newPassword)
	if err != nil {
		return false, err
	}

	if err := s.Queries.UpdatePasswordHashByID(ctx, db.UpdatePasswordHashByIDParams{
		ID:           userID,
		PasswordHash: hashNew,
	}); err != nil {
		return false, err
	}

	if err := s.Queries.DeleteAuthTokenByUserID(ctx, userID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}

	return false, nil
}

func (s *Store) Login(ctx context.Context, username, password string) (rawToken string, expiresAt time.Time, user db.User, err error) {
	user, err = s.Queries.GetUserByUsername(ctx, username)
	if err != nil {
		return "", time.Time{}, db.User{}, ErrInvalidCredentials
	}

	if !auth.CheckPassword(user.PasswordHash, password) {
		return "", time.Time{}, db.User{}, ErrInvalidCredentials
	}

	raw, hash, err := auth.GenerateRawToken()
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

func (s *Store) ExtendTokenExpiry(ctx context.Context, rawToken string) (string, error) {
	hash := auth.HashToken(rawToken)

	token, err := s.Queries.GetAuthTokenByHash(ctx, hash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrTokenInvalidOrExpired
		}
		log.Printf("500 at ExtendTokenExpiry: GetAuthTokenByHash: %v", err)
		return "", err
	}

	if token.ExpiresAt.Time.Before(time.Now()) {
		if err = s.Queries.DeleteAuthToken(ctx, hash); err != nil {
			log.Printf("deleting auth token: %v", err)
		}
		return "", ErrTokenInvalidOrExpired
	}

	newExpiry := time.Now().Add(auth.TokenTTL)
	if err := s.Queries.RefreshAuthTokenExpiry(ctx, db.RefreshAuthTokenExpiryParams{
		ID:        token.ID,
		ExpiresAt: ToPgTimestamptz(newExpiry),
	}); err != nil {
		return "", fmt.Errorf("refreshing auth token: %w", err)
	}

	return token.UserID, nil
}

func (s *Store) Logout(ctx context.Context, rawToken string) error {
	return s.Queries.DeleteAuthToken(ctx, auth.HashToken(rawToken))
}

func (s *Store) QueryUser(ctx context.Context, userID string) (db.User, error) {
	user, err := s.Queries.GetUserByID(ctx, userID)
	if err != nil {
		return db.User{}, fmt.Errorf("Quering user: %w", err)
	}

	return user, nil
}

func (s *Store) UpdateUser(ctx context.Context, userID string, u UpdatedUser) (db.User, error) {
	var clientStuff json.RawMessage
	if u.ClientS != nil {
		clientStuff = *u.ClientS
	}

	user, err := s.Queries.UpdateUserByID(ctx, db.UpdateUserByIDParams{
		ID:            userID,
		Username:      u.Username,
		DisplayName:   u.DisplayName,
		Bio:           ToPgTextPtr(u.Bio),
		Sex:           ToPgTextPtr(u.Sex),
		Birthday:      ToPgDate(u.Birthday),
		LastUpdatedAt: ToPgTimestamptz(time.Now()),
		ClientS:       clientStuff,
	})
	if err != nil {
		return db.User{}, fmt.Errorf("Updating user: %w", err)
	}

	return user, nil
}

func (s *Store) DeleteUserCascade(ctx context.Context, userID string) error {

	return s.WithTx(ctx, func(q *db.Queries) error {
		if err := q.DeleteActivitiesByUser(ctx, userID); err != nil {
			return fmt.Errorf("deleting user's activities: %w", err)
		}
		if err := q.DeleteExercisesByUser(ctx, ToPgText(userID)); err != nil {
			return fmt.Errorf("deleting user's custom exercises: %w", err)
		}
		if err := q.DeleteUser(ctx, userID); err != nil {
			return fmt.Errorf("deleting user: %w", err)
		}
		return nil
	})
}
