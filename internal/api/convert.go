package api

import (
	"encoding/json"
	"log"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/db"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
	"github.com/oapi-codegen/runtime/types"
)

func ToInt32Ptr(p *int) *int32 {
	if p == nil {
		return nil
	}
	v := int32(*p)
	return &v
}

func FromInt64(p int64) *int {
	v := int(p)
	return &v
}

func FromInt32(p int32) *int {
	v := int(p)
	return &v
}

func FromInt32Ptr(p *int32) *int {
	if p == nil {
		return nil
	}
	v := int(*p)
	return &v
}

func TotalWeightToPtr(p int32) *int {
	v := int(p)
	return &v
}

func TotalTimeToPtr(p int64) *int {
	v := int(p)
	return &v
}

func clientS(m *map[string]any) (*json.RawMessage, error) {
	if m == nil {
		return nil, nil
	}
	b, err := json.Marshal(*m)
	if err != nil {
		return nil, err
	}
	raw := json.RawMessage(b)
	return &raw, nil
}

func clientSFromDB(b []byte) *map[string]any {
	if len(b) == 0 {
		return nil
	}
	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		log.Printf("client_s: unmarshal: %v", err)
		return nil
	}
	return &m
}

func dateToPtr(d pgtype.Date) *types.Date {
	if !d.Valid {
		return nil
	}
	v := types.Date{Time: d.Time}
	return &v
}

func toNewActivity(a gen.CreateActivityRequest) (persistence.NewActivity, error) {
	clientS, err := clientS(a.ClientS)
	if err != nil {
		return persistence.NewActivity{}, err
	}

	return persistence.NewActivity{
		ExerciseID:      a.ExerciseId,
		ActivityType:    string(a.ActivityType),
		Reps:            ToInt32Ptr(a.Reps),
		Weight:          a.Weight,
		StartTime:       a.StartTime,
		EndTime:         a.EndTime,
		PerceivedEffort: ToInt32Ptr(a.PerceivedEffort),
		ClientS:         clientS,
	}, nil
}

func toActivityResponse(a db.Activity) gen.ActivityResponse {
	secs := int(a.TotalTime.Microseconds / 1_000_000)
	var weight *float32
	if a.Weight.Valid {
		weight = &a.Weight.Float32
	}

	return gen.ActivityResponse{
		ActivityType:     &a.ActivityType,
		EndTime:          &a.EndTime.Time,
		ExerciseId:       persistence.FromPgTextPtr(a.ExerciseID),
		Id:               &a.ID,
		PerceivedEffort:  FromInt32Ptr(persistence.FromPgInt4Ptr(a.PerceivedEffort)),
		Reps:             FromInt32Ptr(persistence.FromPgInt4Ptr(a.Reps)),
		SortOrder:        FromInt32(a.SortOrder),
		StartTime:        &a.StartTime.Time,
		TotalTimeSeconds: &secs,
		Weight:           weight,
		ClientS:          clientSFromDB(a.ClientS),
	}
}
