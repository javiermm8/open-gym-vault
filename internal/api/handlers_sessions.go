package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
)

// Handles POST /sessions
func (s *Server) CreateSession(w http.ResponseWriter, r *http.Request) {
	var req createSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateCreateSessionRequest(req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	newSession, err := req.toNewSession()
	if err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	session, activities, err := s.store.CreateSessionWithActivities(r.Context(), newSession)
	if err != nil {
		log.Printf("CreateSession: %v", err)
		writeError(w, err)
		return
	}

	resp, err := toSessionResponse(session, activities)
	if err != nil {
		log.Printf("CreateSession: building response: %v", err)
		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)
}

func validateCreateSessionRequest(req createSessionRequest) error {
	if req.UserID == "" {
		return errRequired("user_id")
	}
	if req.SessionType == "" {
		return errRequired("session_type")
	}
	if req.EndTime.Before(req.StartTime) {
		return errInvalid("end_time must not be before start_time")
	}
	for i, a := range req.Activities {
		if err := validateCreateActivityRequest(a); err != nil {
			return fmt.Errorf("activity %d: %w", i, err)
		}
	}
	return nil
}

func validateCreateActivityRequest(a createActivityRequest) error {
	switch a.ActivityType {
	case "exercise":
		if a.ExerciseID == nil {
			return errInvalid("exercise_id is required when activity_type is \"exercise\"")
		}
	case "rest":
		if a.ExerciseID != nil {
			return errInvalid("exercise_id must not be set when activity_type is \"rest\"")
		}
	default:
		return errInvalid(`activity_type must be "exercise" or "rest"`)
	}
	if a.EndTime.Before(a.StartTime) {
		return errInvalid("end_time must not be before start_time")
	}
	return nil
}
