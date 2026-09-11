package api

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/javiermm8/open-gym-vault/internal/db"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

// REQUESTS
type createSessionRequest struct {
	SessionType            string                  `json:"session_type"`
	StartTime              time.Time               `json:"start_time"`
	EndTime                time.Time               `json:"end_time"`
	OverallPerceivedEffort *int32                  `json:"overall_perceived_effort,omitempty"`
	BurnedCals             *int32                  `json:"burned_cals,omitempty"`
	UserNotes              *string                 `json:"user_notes,omitempty"`
	Activities             []createActivityRequest `json:"activities"`
	ClientS                *json.RawMessage        `json:"client_s,omitempty"`
}

type createActivityRequest struct {
	ExerciseID      *string          `json:"exercise_id,omitempty"`
	ActivityType    string           `json:"activity_type"`
	Reps            *int32           `json:"reps,omitempty"`
	Weight          *float32         `json:"weight,omitempty"`
	StartTime       time.Time        `json:"start_time"`
	EndTime         time.Time        `json:"end_time"`
	PerceivedEffort *int32           `json:"perceived_effort,omitempty"`
	ClientS         *json.RawMessage `json:"client_s,omitempty"`
}

type CreateUserRequest struct {
	Username     string           `json:"username"`
	DisplayName  string           `json:"display_name"`
	PasswordHash string           `json:"password_hash"`
	Bio          *string          `json:"bio,omitempty"`
	Sex          *string          `json:"sex,omitempty"`
	Birthday     time.Time        `json:"birthday"`
	ClientS      *json.RawMessage `json:"client_s,omitempty"`
}

type CreateExerciseRequest struct {
	Name             string           `json:"name"`
	AlternativeNames *[]string        `json:"alternative_names,omitempty"`
	Explanation      *string          `json:"explanation,omitempty"`
	ClientS          *json.RawMessage `json:"client_s,omitempty"`
}

type registerRequest struct {
	Username    string `json:"username"`
	DisplayName string `json:"display_name"`
	Password    string `json:"password"`
}

type loginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Turns request into something the internal/persistence layer can work with. Needs the req + a userID(that should come from the token)
func (req createSessionRequest) toNewSession(userID uuid.UUID) (persistence.NewSession, error) {
	activities := make([]persistence.NewActivity, len(req.Activities))
	for i, a := range req.Activities {
		na, err := a.toNewActivity()
		if err != nil {
			return persistence.NewSession{}, fmt.Errorf("activity %d: %w", i, err)
		}
		activities[i] = na
	}

	return persistence.NewSession{
		UserID:                 userID,
		SessionType:            req.SessionType,
		StartTime:              req.StartTime,
		EndTime:                req.EndTime,
		OverallPerceivedEffort: req.OverallPerceivedEffort,
		BurnedCals:             req.BurnedCals,
		UserNotes:              req.UserNotes,
		Activities:             activities,
		ClientS:                req.ClientS,
	}, nil
}

func (req createActivityRequest) toNewActivity() (persistence.NewActivity, error) {
	var exerciseID *uuid.UUID
	if req.ExerciseID != nil {
		id, err := uuid.Parse(*req.ExerciseID)
		if err != nil {
			return persistence.NewActivity{}, fmt.Errorf("invalid exercise_id: %w", err)
		}
		exerciseID = &id
	}

	return persistence.NewActivity{
		ExerciseID:      exerciseID,
		ActivityType:    req.ActivityType,
		Reps:            req.Reps,
		Weight:          req.Weight,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		PerceivedEffort: req.PerceivedEffort,
		ClientS:         req.ClientS,
	}, nil
}

func (req CreateUserRequest) toNewUser() (persistence.NewUser, error) {
	return persistence.NewUser{
		Username:      req.Username,
		DisplayName:   req.DisplayName,
		PasswordHash:  req.PasswordHash,
		Bio:           req.Bio,
		Sex:           req.Sex,
		Birthday:      req.Birthday,
		CreatedAt:     time.Now(),
		LastUpdatedAt: time.Time{},
		ClientS:       req.ClientS,
	}, nil
}

func (req CreateExerciseRequest) toNewExercise(userID uuid.UUID) (persistence.NewExercise, error) {
	return persistence.NewExercise{
		UserID:           userID,
		Name:             req.Name,
		AlternativeNames: req.AlternativeNames,
		Explanation:      req.Explanation,
		ClientS:          req.ClientS,
		CreatedAt:        time.Now(),
		LastUpdatedAt:    time.Time{},
	}, nil
}

// RESPONSES
type sessionResponse struct {
	ID                     string             `json:"id"`
	UserID                 string             `json:"user_id"`
	SessionType            string             `json:"session_type"`
	StartTime              time.Time          `json:"start_time"`
	EndTime                time.Time          `json:"end_time"`
	TotalTimeSeconds       float64            `json:"total_time_seconds"`
	TotalWeight            int32              `json:"total_weight"`
	OverallPerceivedEffort *int32             `json:"overall_perceived_effort,omitempty"`
	BurnedCals             *int32             `json:"burned_cals,omitempty"`
	UserNotes              *string            `json:"user_notes,omitempty"`
	Activities             []activityResponse `json:"activities"`
	ClientS                *json.RawMessage   `json:"client_s,omitempty"`
}

type activityResponse struct {
	ID               string           `json:"id"`
	ExerciseID       *string          `json:"exercise_id,omitempty"`
	ActivityType     string           `json:"activity_type"`
	Reps             *int32           `json:"reps,omitempty"`
	Weight           *float32         `json:"weight,omitempty"`
	SortOrder        int32            `json:"sort_order"`
	StartTime        time.Time        `json:"start_time"`
	EndTime          time.Time        `json:"end_time"`
	TotalTimeSeconds float64          `json:"total_time_seconds"`
	PerceivedEffort  *int32           `json:"perceived_effort,omitempty"`
	ClientS          *json.RawMessage `json:"client_s,omitempty"`
}

type userResponse struct {
	ID          string           `json:"id"`
	Username    string           `json:"username"`
	DisplayName string           `json:"display_name"`
	Bio         *string          `json:"bio,omitempty"`
	Sex         *string          `json:"sex,omitempty"`
	Birthday    time.Time        `json:"birthday"`
	ClientS     *json.RawMessage `json:"client_s,omitempty"`
}

type exerciseFullResponse struct {
	ID               string           `json:"id"`
	UserID           *string          `json:"user_id,omitempty"`
	Name             string           `json:"name"`
	AlternativeNames *[]string        `json:"alternative_names,omitempty"`
	Explanation      *string          `json:"explanation,omitempty"`
	ClientS          *json.RawMessage `json:"client_s,omitempty"`
	CreatedAt        time.Time        `json:"created_at"`
}

type exerciseShortResponse struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type authResponse struct {
	Token       string    `json:"token"`
	ExpiresAt   time.Time `json:"expires_at"`
	UserID      string    `json:"user_id"`
	Username    string    `json:"username"`
	DisplayName string    `json:"display_name"`
}

// Builds the JSON response from the generated db types
func toSessionResponse(session db.Session, activities []db.Activity) (sessionResponse, error) {
	sessionID, err := persistence.FromPgUUID(session.ID)
	if err != nil {
		return sessionResponse{}, fmt.Errorf("session id: %w", err)
	}
	userID, err := persistence.FromPgUUID(session.UserID)
	if err != nil {
		return sessionResponse{}, fmt.Errorf("session user_id: %w", err)
	}

	activityResponses := make([]activityResponse, len(activities))
	for i, a := range activities {
		ar, err := toActivityResponse(a)
		if err != nil {
			return sessionResponse{}, fmt.Errorf("activity %d: %w", i, err)
		}
		activityResponses[i] = ar
	}

	return sessionResponse{
		ID:                     sessionID.String(),
		UserID:                 userID.String(),
		SessionType:            session.SessionType,
		StartTime:              session.StartTime.Time,
		EndTime:                session.EndTime.Time,
		TotalTimeSeconds:       persistence.FromPgInterval(session.TotalTime).Seconds(),
		TotalWeight:            session.TotalWeight,
		OverallPerceivedEffort: persistence.FromPgInt4Ptr(session.OverallPerceivedEffort),
		BurnedCals:             persistence.FromPgInt4Ptr(session.BurnedCals),
		UserNotes:              persistence.FromPgTextPtr(session.UserNotes),
		Activities:             activityResponses,
		ClientS:                (*json.RawMessage)(&session.ClientS),
	}, nil
}

func toActivityResponse(a db.Activity) (activityResponse, error) {
	id, err := persistence.FromPgUUID(a.ID)
	if err != nil {
		return activityResponse{}, fmt.Errorf("id: %w", err)
	}

	var exerciseID *string
	if ptr := persistence.FromPgUUIDPtr(a.ExerciseID); ptr != nil {
		s := ptr.String()
		exerciseID = &s
	}

	return activityResponse{
		ID:               id.String(),
		ExerciseID:       exerciseID,
		ActivityType:     a.ActivityType,
		Reps:             persistence.FromPgInt4Ptr(a.Reps),
		Weight:           persistence.FromPgFloat4Ptr(a.Weight),
		SortOrder:        a.SortOrder,
		StartTime:        a.StartTime.Time,
		EndTime:          a.EndTime.Time,
		TotalTimeSeconds: persistence.FromPgInterval(a.TotalTime).Seconds(),
		PerceivedEffort:  persistence.FromPgInt4Ptr(a.PerceivedEffort),
		ClientS:          (*json.RawMessage)(&a.ClientS),
	}, nil
}

func toUserResponse(user db.User) (userResponse, error) {
	userID, err := persistence.FromPgUUID(user.ID)
	if err != nil {
		return userResponse{}, fmt.Errorf("user id: %w", err)
	}

	return userResponse{
		ID:          userID.String(),
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Bio:         persistence.FromPgTextPtr(user.Bio),
		Sex:         persistence.FromPgTextPtr(user.Sex),
		Birthday:    user.Birthday.Time,
		ClientS:     (*json.RawMessage)(&user.ClientS),
	}, err
}

func toExerciseFullResponse(exercise db.Exercise) (exerciseFullResponse, error) {
	exerciseID, err := persistence.FromPgUUID(exercise.ID)
	if err != nil {
		return exerciseFullResponse{}, fmt.Errorf("exercise id: %w", err)
	}

	var userID string
	if exercise.UserID.String() != "" {
		userUUID, err := persistence.FromPgUUID(exercise.UserID)
		if err != nil {
			return exerciseFullResponse{}, fmt.Errorf("user id(in toExerciseResponse): %w", err)
		}
		userID = userUUID.String()
	} else {
		userID = ""
	}

	return exerciseFullResponse{
		ID:               exerciseID.String(),
		UserID:           &userID,
		Name:             exercise.Name,
		AlternativeNames: &exercise.AlternativeNames,
		Explanation:      persistence.FromPgTextPtr(exercise.Explanation),
		ClientS:          (*json.RawMessage)(&exercise.ClientS),
		CreatedAt:        persistence.FromPgTimestamptz(exercise.CreatedAt),
	}, nil
}

func toExerciseShortResponse(exercise db.Exercise) (exerciseShortResponse, error) {
	exerciseID, err := persistence.FromPgUUID(exercise.ID)
	if err != nil {
		return exerciseShortResponse{}, fmt.Errorf("exercise id: %w", err)
	}

	return exerciseShortResponse{
		ID:   exerciseID.String(),
		Name: exercise.Name,
	}, nil
}
