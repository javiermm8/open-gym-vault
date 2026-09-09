package persistence

import (
	"time"

	"github.com/jackc/pgx/v5/pgtype"
)

func ToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func ToPgDate(t time.Time) pgtype.Date {
	return pgtype.Date{Time: t, InfinityModifier: 0, Valid: true}
}

func ToPgInterval(d time.Duration) pgtype.Interval {
	return pgtype.Interval{Microseconds: d.Microseconds(), Valid: true}
}

func FromPgInterval(i pgtype.Interval) time.Duration {
	return time.Duration(i.Microseconds) * time.Microsecond
}

func ToPgInt4Ptr(v *int32) pgtype.Int4 {
	if v == nil {
		return pgtype.Int4{Valid: false}
	}
	return pgtype.Int4{Int32: *v, Valid: true}
}

func FromPgInt4Ptr(v pgtype.Int4) *int32 {
	if !v.Valid {
		return nil
	}
	i := v.Int32
	return &i
}

func ToPgFloat4Ptr(v *float32) pgtype.Float4 {
	if v == nil {
		return pgtype.Float4{Valid: false}
	}
	return pgtype.Float4{Float32: *v, Valid: true}
}

func FromPgFloat4Ptr(v pgtype.Float4) *float32 {
	if !v.Valid {
		return nil
	}
	f := v.Float32
	return &f
}

func ToPgTextPtr(v *string) pgtype.Text {
	if v == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *v, Valid: true}
}

func FromPgTextPtr(v pgtype.Text) *string {
	if !v.Valid {
		return nil
	}
	s := v.String
	return &s
}

func FromPgTimestamptz(t pgtype.Timestamptz) time.Time {
	if !t.Valid {
		return time.Time{}
	}
	return t.Time
}
