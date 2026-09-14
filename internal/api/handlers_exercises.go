package api

import (
	"context"
	"errors"
	"log"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) CreateExercise(ctx context.Context, r gen.CreateExerciseRequestObject) (gen.CreateExerciseResponseObject, error) {
	if r.Body.Name == "" {
		return gen.CreateExercise400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "name is required",
			},
		}, nil
	}

	clientS, err := clientS(r.Body.ClientS)
	if err != nil {
		return gen.CreateExercise400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "invalid client_s",
			},
		}, nil
	}

	exercise, err := s.store.CreateExercise(ctx, persistence.NewExercise{
		UserID:           AuthenticateUserID(ctx),
		Name:             r.Body.Name,
		AlternativeNames: r.Body.AlternativeNames,
		Explanation:      r.Body.Explanation,
		ClientS:          clientS,
		CreatedAt:        time.Now(),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23514":
				return gen.CreateExercise400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: "invalid data: " + pgErr.Message,
					},
				}, nil
			case "23503":
				return gen.CreateExercise400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: "references a resource that does not exist",
					},
				}, nil
			case "23505":
				return gen.CreateExercise409JSONResponse{
					Error: "A custom exercise with that name already exists",
				}, nil
			}
		}
		log.Printf("500 at CreateExercise: %v", err)
		return nil, err
	}

	return gen.CreateExercise201JSONResponse{
		Id:               &exercise.ID,
		UserId:           &exercise.UserID.String,
		Name:             &exercise.Name,
		AlternativeNames: &exercise.AlternativeNames,
		Explanation:      &exercise.Explanation.String,
		CreatedAt:        &exercise.CreatedAt.Time,
		ClientS:          clientSFromDB(exercise.ClientS),
	}, nil
}

func (s *Server) ListExercises(ctx context.Context, r gen.ListExercisesRequestObject) (gen.ListExercisesResponseObject, error) {
	exercises, err := s.store.QueryExercises(ctx, AuthenticateUserID(ctx))
	if err != nil {
		log.Printf("500 at ListExercises: %v", err)
		return nil, err
	}

	if len(exercises) <= 0 {
		return gen.ListExercises204Response{}, err
	}

	full := r.Params.Full != nil && *r.Params.Full

	exerciseResponses := make([]gen.ExerciseResponse, len(exercises))
	for i, n := range exercises {
		exerciseResponses[i] = gen.ExerciseResponse{
			ClientS: clientSFromDB(n.ClientS),
			Id:      &n.ID,
			Name:    &n.Name,
		}
		if full {
			exerciseResponses[i].AlternativeNames = &n.AlternativeNames
			exerciseResponses[i].CreatedAt = &n.CreatedAt.Time
			exerciseResponses[i].Explanation = &n.Explanation.String
			exerciseResponses[i].UserId = &n.UserID.String
		}
	}

	return gen.ListExercises200JSONResponse(exerciseResponses), nil
}

func (s *Server) GetExerciseById(ctx context.Context, r gen.GetExerciseByIdRequestObject) (gen.GetExerciseByIdResponseObject, error) {
	if r.Id == "" {
		return gen.GetExerciseById400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id is required",
			},
		}, nil
	}
	if utf8.RuneCountInString(r.Id) != 36 {
		return gen.GetExerciseById400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id must be a valid exercise id",
			},
		}, nil
	}

	exercise, err := s.store.QueryExecise(ctx, r.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.GetExerciseById403JSONResponse{
				Error: "Forbidden",
			}, nil
		}
		log.Printf("500 at GetExerciseById: %v", err)
		return nil, err
	}

	if persistence.FromPgText(exercise.UserID) != "" {
		if AuthenticateUserID(ctx) != persistence.FromPgText(exercise.UserID) {
			return gen.GetExerciseById403JSONResponse{
				Error: "Forbidden",
			}, nil
		}
	}

	createdTime := persistence.FromPgTimestamptz(exercise.CreatedAt)

	return gen.GetExerciseById200JSONResponse{
		Id:               &exercise.ID,
		UserId:           persistence.FromPgTextPtr(exercise.UserID),
		Name:             &exercise.Name,
		AlternativeNames: &exercise.AlternativeNames,
		Explanation:      persistence.FromPgTextPtr(exercise.Explanation),
		ClientS:          clientSFromDB(exercise.ClientS),
		CreatedAt:        &createdTime,
	}, nil
}
