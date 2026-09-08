package persistence

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/javiermm8/open-gym-vault/internal/db"
)

type NewActivity struct {
	ExerciseID      *uuid.UUID // nil for rest activities
	ActivityType    string     // "exercise" or "rest"
	Reps            *int32
	Weight          *float32 // kg
	StartTime       time.Time
	EndTime         time.Time
	PerceivedEffort *int32
}

type NewSession struct {
	UserID                 uuid.UUID
	SessionType            string
	StartTime              time.Time
	EndTime                time.Time
	OverallPerceivedEffort *int32
	BurnedCals             *int32
	UserNotes              *string
	Activities             []NewActivity
}

func (s *Store) CreateSessionWithActivities(ctx context.Context, in NewSession) (db.Session, []db.Activity, error) {
	var session db.Session
	var activities []db.Activity

	err := s.WithTx(ctx, func(q *db.Queries) error {
		totalWeight := computeTotalWeight(in.Activities)

		var err error
		session, err = q.CreateSession(ctx, db.CreateSessionParams{
			UserID:                 ToPgUUID(in.UserID),
			SessionType:            in.SessionType,
			StartTime:              ToPgTimestamptz(in.StartTime),
			EndTime:                ToPgTimestamptz(in.EndTime),
			TotalTime:              ToPgInterval(in.EndTime.Sub(in.StartTime)),
			TotalWeight:            totalWeight,
			OverallPerceivedEffort: ToPgInt4Ptr(in.OverallPerceivedEffort),
			BurnedCals:             ToPgInt4Ptr(in.BurnedCals),
			UserNotes:              ToPgTextPtr(in.UserNotes),
		})
		if err != nil {
			return fmt.Errorf("creating session: %w", err)
		}

		activities = make([]db.Activity, 0, len(in.Activities))
		for i, a := range in.Activities {
			activity, err := q.CreateActivity(ctx, db.CreateActivityParams{
				SessionID:       session.ID,
				ExerciseID:      ToPgUUIDPtr(a.ExerciseID),
				ActivityType:    a.ActivityType,
				Reps:            ToPgInt4Ptr(a.Reps),
				Weight:          ToPgFloat4Ptr(a.Weight),
				SortOrder:       int32(i),
				StartTime:       ToPgTimestamptz(a.StartTime),
				EndTime:         ToPgTimestamptz(a.EndTime),
				TotalTime:       ToPgInterval(a.EndTime.Sub(a.StartTime)),
				PerceivedEffort: ToPgInt4Ptr(a.PerceivedEffort),
			})
			if err != nil {
				return fmt.Errorf("creating activity %d: %w", i, err)
			}
			activities = append(activities, activity)
		}

		return nil
	})
	if err != nil {
		return db.Session{}, nil, err
	}

	return session, activities, nil
}

func computeTotalWeight(activities []NewActivity) int32 {
	var total float32
	for _, a := range activities {
		if a.Weight == nil || a.Reps == nil {
			continue
		}
		total += *a.Weight * float32(*a.Reps)
	}
	return int32(total)
}
