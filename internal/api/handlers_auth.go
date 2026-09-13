package api

import (
	"context"
	"errors"
	"log"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) RegisterUser(ctx context.Context, r gen.RegisterUserRequestObject) (gen.RegisterUserResponseObject, error) {
	if r.Body.Username == "" || r.Body.DisplayName == "" || r.Body.Password == "" {
		return gen.RegisterUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "username, display_name and password are required",
			},
		}, nil
	}

	user, err := s.store.Register(ctx, r.Body.Username, r.Body.DisplayName, r.Body.Password)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return gen.RegisterUser409JSONResponse{
				Error: "username already in use",
			}, nil
		}
		return nil, err
	}

	return gen.RegisterUser201JSONResponse{
		Id:          &user.ID,
		Username:    &user.Username,
		DisplayName: &user.DisplayName,
	}, nil
}

func (s *Server) LoginUser(ctx context.Context, r gen.LoginUserRequestObject) (gen.LoginUserResponseObject, error) {

	if r.Body.Username == "" || r.Body.Password == "" {
		return gen.LoginUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "username and password are required",
			},
		}, nil
	}

	rawToken, expiresAt, user, err := s.store.Login(ctx, r.Body.Username, r.Body.Password)
	if err != nil {
		if errors.Is(err, persistence.ErrInvalidCredentials) {
			return gen.LoginUser401JSONResponse{
				Error: "invalid username or password",
			}, nil
		}
		log.Printf("Login: %v", err)
		return nil, err
	}

	return loginResponse{
		LoginUser200JSONResponse: gen.LoginUser200JSONResponse{
			Token: &rawToken, ExpiresAt: &expiresAt, UserId: &user.ID, Username: &user.Username,
		},
		token:     rawToken,
		expiresAt: expiresAt,
	}, nil
}

func (s *Server) LogoutUser(ctx context.Context, r gen.LogoutUserRequestObject) (gen.LogoutUserResponseObject, error) {
	token, ok := TokenFromContext(ctx)
	if !ok {
		return gen.LogoutUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "no token provided",
			},
		}, nil
	}
	if err := s.store.Logout(ctx, token); err != nil &&
		!errors.Is(err, persistence.ErrTokenInvalidOrExpired) {
		return nil, err
	}
	return logoutResponse{}, nil
}
