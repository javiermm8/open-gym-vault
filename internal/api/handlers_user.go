package api

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"unicode/utf8"
)

const maxUsernameLength = 10
const maxDisplayNameLength = 40

func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
		return
	}

	if err := validateCreateUserRequest(req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	newUser, err := req.toNewUser()
	if err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := s.store.CreateUser(r.Context(), newUser)
	if err != nil {
		log.Printf("CreateUser: %v", err)
		writeError(w, err)
		return
	}

	resp, err := toUserResponse(user)
	if err != nil {
		log.Printf("CreateUser: building response: %v", err)
		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusCreated, resp)

}

func validateCreateUserRequest(req CreateUserRequest) error {
	if req.Username == "" {
		return errRequired("username")
	}
	if req.DisplayName == "" {
		return errRequired("display_name")
	}
	if req.PasswordHash == "" {
		return errRequired("password_hash")
	}

	uLength := utf8.RuneCountInString(req.Username)
	dLength := utf8.RuneCountInString(req.DisplayName)

	if uLength > maxUsernameLength {
		return fmt.Errorf("Username too long. Length: %d. Max length: %d", uLength, maxUsernameLength)
	}

	if dLength > maxDisplayNameLength {
		return fmt.Errorf("Display name too long. Length: %d. Max length: %d", dLength, maxDisplayNameLength)
	}

	return nil
}
