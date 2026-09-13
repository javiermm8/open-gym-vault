package api

import (
	"context"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) GetCurrentUser(ctx context.Context, r gen.GetCurrentUserRequestObject) (gen.GetCurrentUserResponseObject, error) {
	user, err := s.store.QueryUser(ctx, AuthenticateUserID(ctx))
	if err != nil {
		return nil, err
	}

	return gen.GetCurrentUser200JSONResponse{
		Id:          &user.ID,
		Username:    &user.Username,
		DisplayName: &user.DisplayName,
		Bio:         persistence.FromPgTextPtr(user.Bio),
		Sex:         persistence.FromPgTextPtr(user.Sex),
		Birthday:    dateToPtr(user.Birthday),
		ClientS:     clientSFromDB(user.ClientS),
	}, nil
}

func (s *Server) UpdateCurrentUser(ctx context.Context, r gen.UpdateCurrentUserRequestObject) (gen.UpdateCurrentUserResponseObject, error) {
	clientS, err := clientS(r.Body.ClientS)
	if err != nil {
		return gen.UpdateCurrentUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "invalid client_s",
			},
		}, nil
	}

	var birthday time.Time
	if r.Body.Birthday != nil {
		birthday = r.Body.Birthday.Time
	}
	var username string
	if r.Body.Username != nil {
		username = *r.Body.Username
	}
	var displayName string
	if r.Body.DisplayName != nil {
		displayName = *r.Body.DisplayName
	}

	user, err := s.store.UpdateUser(ctx, AuthenticateUserID(ctx), persistence.UpdatedUser{
		Username:    username,
		DisplayName: displayName,
		Bio:         r.Body.Bio,
		Sex:         r.Body.Sex,
		Birthday:    birthday,
		ClientS:     clientS,
	})

	if err != nil {
		return nil, err
	}

	if user.Birthday.Time.IsZero() {
		return gen.UpdateCurrentUser200JSONResponse{
			Id:          &user.ID,
			Username:    &user.Username,
			DisplayName: &user.DisplayName,
			Bio:         persistence.FromPgTextPtr(user.Bio),
			Sex:         persistence.FromPgTextPtr(user.Sex),
			ClientS:     clientSFromDB(user.ClientS),
		}, nil
	} else {
		return gen.UpdateCurrentUser200JSONResponse{
			Id:          &user.ID,
			Username:    &user.Username,
			DisplayName: &user.DisplayName,
			Bio:         persistence.FromPgTextPtr(user.Bio),
			Sex:         persistence.FromPgTextPtr(user.Sex),
			Birthday:    dateToPtr(user.Birthday),
			ClientS:     clientSFromDB(user.ClientS),
		}, nil
	}
}
