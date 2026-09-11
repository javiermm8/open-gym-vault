package api

import (
	"encoding/json"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

func (s *Server) CreateExercise(w http.ResponseWriter, r *http.Request) {
	var req CreateExerciseRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateCreateExerciseRequest(req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	newExercise, err := req.toNewExercise(AuthenticateUserID(r))
	if err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	exercise, err := s.store.CreateExercise(r.Context(), newExercise)
	if err != nil {
		log.Printf("CreateExercise: %v", err)
		writeError(w, err)
		return
	}

	resp := toExerciseFullResponse(exercise)

	writeJSON(w, http.StatusCreated, resp)
}

func validateCreateExerciseRequest(req CreateExerciseRequest) error {
	if req.Name == "" {
		return errRequired("name")
	}

	return nil
}

func (s *Server) GetExercise(w http.ResponseWriter, r *http.Request) {
	exerciseID := r.PathValue("id")

	if err := validateGetExerciseRequest(exerciseID); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	exercise, err := s.store.QueryExecise(r.Context(), exerciseID)
	if err != nil {
		log.Printf("GetExercise: %v", err)
		writeError(w, err)
		return
	}

	if persistence.FromPgText(exercise.UserID) != "" {
		if AuthenticateUserID(r) != persistence.FromPgText(exercise.UserID) {
			writeErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	resp := toExerciseFullResponse(exercise)

	writeJSON(w, http.StatusOK, resp)
}

func validateGetExerciseRequest(id string) error {
	if id == "" {
		return errRequired("id")
	}
	if utf8.RuneCountInString(id) != 36 {
		return errInvalid("id must be a valid exercise id")
	}

	return nil
}

func (s Server) ListExercises(w http.ResponseWriter, r *http.Request) {
	exercises, err := s.store.QueryExercises(r.Context(), AuthenticateUserID(r))
	if err != nil {
		log.Printf("ListExercises: %v", err)
		writeError(w, err)
		return
	}

	if len(exercises) <= 0 {
		writeJSON(w, http.StatusNoContent, nil)
		return
	}

	full := r.URL.Query().Get("full")
	if full == "true" {
		var exerciseResponses []exerciseFullResponse
		for _, n := range exercises {
			exerciseResponses = append(exerciseResponses, toExerciseFullResponse(n))
		}
		writeJSON(w, http.StatusOK, exerciseResponses)
	} else if full == "false" || full == "" {
		var exerciseResponses []exerciseShortResponse
		for _, n := range exercises {
			resp := toExerciseShortResponse(n)

			exerciseResponses = append(exerciseResponses, resp)
		}
		writeJSON(w, http.StatusOK, exerciseResponses)
	} else {
		writeErrorMessage(w, http.StatusBadRequest, "Available flags: full=true/false")
		return
	}
}
