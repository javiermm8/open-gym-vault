package api

import (
	"context"
	"errors"
	"log"
	"strconv"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) CreateSession(ctx context.Context, r gen.CreateSessionRequestObject) (gen.CreateSessionResponseObject, error) {

	// Validate request
	if r.Body.SessionType == "" {
		return gen.CreateSession400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "session_type is required",
			},
		}, nil
	}
	if r.Body.EndTime.Before(r.Body.StartTime) {
		return gen.CreateSession400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "end_time must not be before start_time",
			},
		}, nil
	}
	for i, a := range r.Body.Activities {
		switch a.ActivityType {
		case "exercise":
			if a.ExerciseId == nil {
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: `activity ` + strconv.Itoa(i) + `: exercise_id is required when activity_type is "exercise"`,
					},
				}, nil
			}
		case "rest":
			if a.ExerciseId != nil {
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: `activity ` + strconv.Itoa(i) + `: exercise_id must not be set when activity_type is "rest"`,
					},
				}, nil
			}
			if a.Reps != nil {
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: `activity ` + strconv.Itoa(i) + `: reps must not be set when activity_type is "rest"`,
					},
				}, nil
			}
			if a.Weight != nil {
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: `activity ` + strconv.Itoa(i) + `: weight must not be set when activity_type is "rest"`,
					},
				}, nil
			}
		case "other":
		default:
			return gen.CreateSession400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{
					Error: `activity ` + strconv.Itoa(i) + `: activity_type must be "exercise", "rest" or other`,
				},
			}, nil
		}

		if a.EndTime.Before(a.StartTime) {
			return gen.CreateSession400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{
					Error: "activity " + strconv.Itoa(i) + ": end_time must not be before start_time",
				},
			}, nil
		}
	}

	activities := make([]persistence.NewActivity, len(r.Body.Activities))
	for i, a := range r.Body.Activities {
		var err error
		activities[i], err = toNewActivity(a)
		if err != nil {
			return gen.CreateSession400JSONResponse{
				BadRequestJSONResponse: gen.BadRequestJSONResponse{
					Error: "invalid client_s in activity " + strconv.Itoa(i),
				},
			}, nil
		}
	}

	clientS, err := clientS(r.Body.ClientS)
	if err != nil {
		return gen.CreateSession400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "invalid client_s",
			},
		}, nil
	}

	session, activitiesDB, err := s.store.CreateSessionWithActivities(ctx, persistence.NewSession{
		UserID:                 AuthenticateUserID(ctx),
		SessionType:            r.Body.SessionType,
		StartTime:              r.Body.StartTime,
		EndTime:                r.Body.EndTime,
		OverallPerceivedEffort: ToInt32Ptr(r.Body.OverallPerceivedEffort),
		BurnedCals:             ToInt32Ptr(r.Body.BurnedCals),
		UserNotes:              r.Body.UserNotes,
		Activities:             activities,
		ClientS:                clientS,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23514":
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: "invalid data: " + pgErr.Message,
					},
				}, nil
			case "23503":
				return gen.CreateSession400JSONResponse{
					BadRequestJSONResponse: gen.BadRequestJSONResponse{
						Error: "references a resource that does not exist",
					},
				}, nil
			default:
				return nil, err
			}
		}
		log.Printf("500 at CreateSession: %v", err)
		return nil, err
	}

	var activitiesOut []gen.ActivityResponse
	for _, n := range activitiesDB {
		activity := toActivityResponse(n)
		activitiesOut = append(activitiesOut, activity)
	}

	return gen.CreateSession201JSONResponse{
		Id:                     &session.ID,
		UserId:                 &session.UserID,
		StartTime:              &session.StartTime.Time,
		EndTime:                &session.EndTime.Time,
		TotalTimeSeconds:       TotalTimeToPtr(int64(persistence.FromPgInterval(session.TotalTime) / time.Second)),
		TotalWeight:            TotalWeightToPtr(session.TotalWeight),
		OverallPerceivedEffort: FromInt32Ptr(persistence.FromPgInt4Ptr(session.OverallPerceivedEffort)),
		BurnedCals:             FromInt32Ptr(persistence.FromPgInt4Ptr(session.BurnedCals)),
		UserNotes:              persistence.FromPgTextPtr(session.UserNotes),
		Activities:             &activitiesOut,
		ClientS:                clientSFromDB(session.ClientS),
	}, nil
}

func (s *Server) ListSessions(ctx context.Context, r gen.ListSessionsRequestObject) (gen.ListSessionsResponseObject, error) {
	sessions, err := s.store.QuerySessions(ctx, AuthenticateUserID(ctx))
	if err != nil {
		log.Printf("500 at ListSessions: %v", err)
		return nil, err
	}

	if len(sessions) <= 0 {
		return gen.ListSessions204Response{}, err
	}

	full := r.Params.Full != nil && *r.Params.Full

	var activitiesBySession map[string][]gen.ActivityResponse
	if full {
		activitiesDB, err := s.store.QueryActivitiesByUser(ctx, AuthenticateUserID(ctx))
		if err != nil {
			log.Printf("500 at ListSessions: %v", err)
			return nil, err
		}
		activitiesBySession = make(map[string][]gen.ActivityResponse, len(sessions))
		for _, a := range activitiesDB {
			activitiesBySession[a.SessionID] = append(activitiesBySession[a.SessionID], toActivityResponse(a))
		}
	}

	sessionResponses := make([]gen.SessionResponse, len(sessions))
	for i, n := range sessions {
		sessionResponses[i] = gen.SessionResponse{
			Id:          &n.ID,
			SessionType: &n.SessionType,
			ClientS:     clientSFromDB(n.ClientS),
		}
		if full {
			activities, ok := activitiesBySession[n.ID]
			if !ok {
				activities = make([]gen.ActivityResponse, 0)
			}
			sessionResponses[i].Activities = &activities
			sessionResponses[i].EndTime = &n.EndTime.Time
			sessionResponses[i].OverallPerceivedEffort = FromInt32Ptr(persistence.FromPgInt4Ptr(n.OverallPerceivedEffort))
			sessionResponses[i].StartTime = &n.StartTime.Time
			sessionResponses[i].TotalTimeSeconds = TotalTimeToPtr(int64(persistence.FromPgInterval(n.TotalTime) / time.Second))
			sessionResponses[i].TotalWeight = FromInt32(n.TotalWeight)
			sessionResponses[i].UserId = &n.UserID
			sessionResponses[i].UserNotes = persistence.FromPgTextPtr(n.UserNotes)
		}
	}

	return gen.ListSessions200JSONResponse(sessionResponses), nil

}

func (s *Server) GetSessionById(ctx context.Context, r gen.GetSessionByIdRequestObject) (gen.GetSessionByIdResponseObject, error) {
	if r.Id == "" {
		return gen.GetSessionById400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id is required",
			},
		}, nil
	}
	if utf8.RuneCountInString(r.Id) != 36 {
		return gen.GetSessionById400JSONResponse{
			BadRequestJSONResponse: gen.BadRequestJSONResponse{
				Error: "id must be a valid session id",
			},
		}, nil
	}

	session, activities, err := s.store.QuerySession(ctx, r.Id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return gen.GetSessionById403JSONResponse{
				Error: "forbidden",
			}, nil
		}
		log.Printf("500 at GetSessionById: %v", err)
		return nil, err
	}

	if AuthenticateUserID(ctx) != session.UserID {
		return gen.GetSessionById403JSONResponse{
			Error: "forbidden",
		}, nil
	}

	var activitiesOut []gen.ActivityResponse
	for _, n := range activities {
		activity := toActivityResponse(n)
		activitiesOut = append(activitiesOut, activity)
	}

	return gen.GetSessionById200JSONResponse{
		Activities:             &activitiesOut,
		BurnedCals:             FromInt32Ptr(persistence.FromPgInt4Ptr(session.BurnedCals)),
		EndTime:                &session.EndTime.Time,
		Id:                     &session.ID,
		OverallPerceivedEffort: FromInt32Ptr(persistence.FromPgInt4Ptr(session.OverallPerceivedEffort)),
		SessionType:            &session.SessionType,
		StartTime:              &session.StartTime.Time,
		TotalTimeSeconds:       TotalTimeToPtr(int64(persistence.FromPgInterval(session.TotalTime) / time.Second)),
		TotalWeight:            TotalWeightToPtr(session.TotalWeight),
		UserId:                 &session.UserID,
		UserNotes:              persistence.FromPgTextPtr(session.UserNotes),
		ClientS:                clientSFromDB(session.ClientS),
	}, nil
}
