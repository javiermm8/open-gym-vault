package api

import (
	"encoding/json"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/google/uuid"
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

	newExercise, err := req.toNewExercise()
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

	resp, err := toExerciseResponse(exercise)
	if err != nil {
		log.Printf("CreateExercise: building response: %v", err)
		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
		return
	}
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

	if exercise.UserID.String() != "" {
		userID, err := persistence.FromPgUUID(exercise.UserID)
		if err != nil {
			log.Printf("GetExercise: FromPgUUID: %v", err)
			writeError(w, err)
			return
		}

		if AuthenticateUserID(r) != userID {
			writeErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
			return
		}
	}

	resp, err := toExerciseResponse(exercise)
	if err != nil {
		log.Printf("GetExercise: building response: %v", err)
		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)

}

func validateGetExerciseRequest(id string) error {
	if id == "" {
		return errRequired("id")
	}
	if utf8.RuneCountInString(id) != 36 {
		return errInvalid("id must be a valid exercise id")
	}
	if err := uuid.Validate(id); err != nil {
		return errInvalid("id must be a valid exercise id")
	}

	return nil
}
