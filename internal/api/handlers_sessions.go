package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"
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

	newSession := req.toNewSession(AuthenticateUserID(r))

	session, activities, err := s.store.CreateSessionWithActivities(r.Context(), newSession)
	if err != nil {
		log.Printf("CreateSession: %v", err)
		writeError(w, err)
		return
	}

	resp := toSessionResponse(session, activities)

	writeJSON(w, http.StatusCreated, resp)
}

func validateCreateSessionRequest(req createSessionRequest) error {
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
			return errInvalid(`exercise_id is required when activity_type is "exercise"`)
		}
	case "rest":
		if a.ExerciseID != nil {
			return errInvalid(`exercise_id must not be set when activity_type is "rest"`)
		}
		if a.Reps != nil {
			return errInvalid(`reps must not be set when activity_type is "rest"`)
		}
		if a.Weight != nil {
			return errInvalid(`weight must not be set when activity_type is "rest"`)
		}
	case "other":
	default:
		return errInvalid(`activity_type must be "exercise" or "rest"`)
	}
	if a.EndTime.Before(a.StartTime) {
		return errInvalid("end_time must not be before start_time")
	}
	return nil
}

func (s *Server) GetSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("id")

	if err := validateGetSessionRequest(sessionID); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	session, activities, err := s.store.GetSession(r.Context(), sessionID)
	if err != nil {
		log.Printf("GetSession: %v", err)
		writeError(w, err)
		return
	}

	if AuthenticateUserID(r) != session.UserID {
		writeErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp := toSessionResponse(session, activities)

	writeJSON(w, http.StatusOK, resp)

}

func validateGetSessionRequest(id string) error {
	if id == "" {
		return errRequired("id")
	}
	if utf8.RuneCountInString(id) != 36 {
		return errInvalid("id must be a valid session id")
	}

	return nil
}

func (s *Server) ListSessions(w http.ResponseWriter, r *http.Request) {
	sessions, err := s.store.QuerySessions(r.Context(), AuthenticateUserID(r))
	if err != nil {
		log.Printf("List Sessions: %v", err)
		writeError(w, err)
		return
	}

	if len(sessions) <= 0 {
		writeJSON(w, http.StatusNoContent, nil)
		return
	}

	full := r.URL.Query().Get("full")
	if full == "true" {
		var resp []sessionResponse
		for _, n := range sessions {
			activities, err := s.store.QueryActivities(r.Context(), n.ID)
			if err != nil {
				log.Printf("List Activity: %v", err)
				writeError(w, err)
				return
			}
			resp = append(resp, toSessionResponse(n, activities))
		}
		writeJSON(w, http.StatusOK, resp)
	} else if full == "false" || full == "" {
		var resp []listSessionShortResponse
		for _, n := range sessions {
			resp = append(resp, toListShortSessionResponse(n))
		}
		writeJSON(w, http.StatusOK, resp)
	} else {
		writeErrorMessage(w, http.StatusBadRequest, "Available flags: full=true/false")
		return
	}
}
