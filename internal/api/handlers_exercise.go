package api

import (
	"encoding/json"
	"log"
	"net/http"
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
