package api

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) GetCurrentUser(ctx context.Context, r gen.GetCurrentUserRequestObject) (gen.GetCurrentUserResponseObject, error) {
	user, err := s.store.QueryUser(ctx, AuthenticateUserID(ctx))
	if err != nil {
		log.Printf("500 at GetCurrentUser: %v", err)
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
	clientSr, err := clientS(r.Body.ClientS)
	if err != nil {
		return gen.UpdateCurrentUser400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "invalid client_s",
			},
		}, nil
	}

	userDB, err := s.store.QueryUser(ctx, AuthenticateUserID(ctx))
	if err != nil {
		log.Printf("500 at GetCurrentUser: %v", err)
		return nil, err
	}

	var birthday time.Time
	if r.Body.Birthday != nil {
		birthday = r.Body.Birthday.Time
	} else if dateToPtr(userDB.Birthday) != nil {
		birthday = userDB.Birthday.Time
	}
	var username string
	if r.Body.Username != nil {
		username = *r.Body.Username
	} else {
		username = userDB.Username
	}
	var displayName string
	if r.Body.DisplayName != nil {
		displayName = *r.Body.DisplayName
	} else {
		displayName = userDB.DisplayName
	}
	var bio string
	if r.Body.Bio != nil {
		bio = *r.Body.Bio
	} else {
		bio = userDB.Bio.String
	}
	var sex string
	if r.Body.Sex != nil {
		sex = *r.Body.Sex
	} else {
		sex = userDB.Sex.String
	}
	var clientS json.RawMessage
	if clientSr != nil {
		clientS = *clientSr
	} else {
		clientS = userDB.ClientS
	}

	user, err := s.store.UpdateUser(ctx, AuthenticateUserID(ctx), persistence.UpdatedUser{
		Username:    username,
		DisplayName: displayName,
		Bio:         &bio,
		Sex:         &sex,
		Birthday:    birthday,
		ClientS:     &clientS,
	})

	if err != nil {
		log.Printf("500 at UpdateCurrentUser: %v", err)
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
