package api

import (
	"encoding/json"
	"time"

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
func (req createSessionRequest) toNewSession(userID string) persistence.NewSession {
	activities := make([]persistence.NewActivity, len(req.Activities))
	for i, a := range req.Activities {
		na := a.toNewActivity()
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
	}
}

func (req createActivityRequest) toNewActivity() persistence.NewActivity {
	return persistence.NewActivity{
		ExerciseID:      req.ExerciseID,
		ActivityType:    req.ActivityType,
		Reps:            req.Reps,
		Weight:          req.Weight,
		StartTime:       req.StartTime,
		EndTime:         req.EndTime,
		PerceivedEffort: req.PerceivedEffort,
		ClientS:         req.ClientS,
	}
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

func (req CreateExerciseRequest) toNewExercise(userID string) (persistence.NewExercise, error) {
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
func toSessionResponse(session db.Session, activities []db.Activity) sessionResponse {
	activityResponses := make([]activityResponse, len(activities))
	for i, a := range activities {
		ar := toActivityResponse(a)
		activityResponses[i] = ar
	}

	return sessionResponse{
		ID:                     session.ID,
		UserID:                 session.UserID,
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
	}
}

func toActivityResponse(a db.Activity) activityResponse {
	return activityResponse{
		ID:               a.ID,
		ExerciseID:       persistence.FromPgTextPtr(a.ExerciseID),
		ActivityType:     a.ActivityType,
		Reps:             persistence.FromPgInt4Ptr(a.Reps),
		Weight:           persistence.FromPgFloat4Ptr(a.Weight),
		SortOrder:        a.SortOrder,
		StartTime:        a.StartTime.Time,
		EndTime:          a.EndTime.Time,
		TotalTimeSeconds: persistence.FromPgInterval(a.TotalTime).Seconds(),
		PerceivedEffort:  persistence.FromPgInt4Ptr(a.PerceivedEffort),
		ClientS:          (*json.RawMessage)(&a.ClientS),
	}
}

func toUserResponse(user db.User) userResponse {
	return userResponse{
		ID:          user.ID,
		Username:    user.Username,
		DisplayName: user.DisplayName,
		Bio:         persistence.FromPgTextPtr(user.Bio),
		Sex:         persistence.FromPgTextPtr(user.Sex),
		Birthday:    user.Birthday.Time,
		ClientS:     (*json.RawMessage)(&user.ClientS),
	}
}

func toExerciseFullResponse(exercise db.Exercise) exerciseFullResponse {
	return exerciseFullResponse{
		ID:               exercise.ID,
		UserID:           persistence.FromPgTextPtr(exercise.UserID),
		Name:             exercise.Name,
		AlternativeNames: &exercise.AlternativeNames,
		Explanation:      persistence.FromPgTextPtr(exercise.Explanation),
		ClientS:          (*json.RawMessage)(&exercise.ClientS),
		CreatedAt:        persistence.FromPgTimestamptz(exercise.CreatedAt),
	}
}

func toExerciseShortResponse(exercise db.Exercise) exerciseShortResponse {
	return exerciseShortResponse{
		ID:   exercise.ID,
		Name: exercise.Name,
	}
}
