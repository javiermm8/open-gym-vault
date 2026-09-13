package api

import (
	"context"
	"errors"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
)

var errNotImplemented = errors.New("not implemented")

func (s *Server) LogoutUser(ctx context.Context, r gen.LogoutUserRequestObject) (gen.LogoutUserResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ListExercises(ctx context.Context, r gen.ListExercisesRequestObject) (gen.ListExercisesResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetExerciseById(ctx context.Context, r gen.GetExerciseByIdRequestObject) (gen.GetExerciseByIdResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) CreateExercise(ctx context.Context, r gen.CreateExerciseRequestObject) (gen.CreateExerciseResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) CreateSession(ctx context.Context, r gen.CreateSessionRequestObject) (gen.CreateSessionResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) ListSessions(ctx context.Context, r gen.ListSessionsRequestObject) (gen.ListSessionsResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetSessionById(ctx context.Context, r gen.GetSessionByIdRequestObject) (gen.GetSessionByIdResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) GetCurrentUser(ctx context.Context, r gen.GetCurrentUserRequestObject) (gen.GetCurrentUserResponseObject, error) {
	return nil, errNotImplemented
}

func (s *Server) UpdateCurrentUser(ctx context.Context, r gen.UpdateCurrentUserRequestObject) (gen.UpdateCurrentUserResponseObject, error) {
	return nil, errNotImplemented
}
