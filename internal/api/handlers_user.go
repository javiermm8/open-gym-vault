package api

import (
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

const maxUsernameLength = 10
const maxDisplayNameLength = 40

// func (s *Server) CreateUser(w http.ResponseWriter, r *http.Request) {
// 	var req CreateUserRequest
// 	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
// 		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
// 		return
// 	}

// 	if err := validateCreateUserRequest(req); err != nil {
// 		writeErrorMessage(w, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	newUser, err := req.toNewUser()
// 	if err != nil {
// 		writeErrorMessage(w, http.StatusBadRequest, err.Error())
// 		return
// 	}

// 	user, err := s.store.CreateUser(r.Context(), newUser)
// 	if err != nil {
// 		log.Printf("CreateUser: %v", err)
// 		writeError(w, err)
// 		return
// 	}

// 	resp, err := toUserResponse(user)
// 	if err != nil {
// 		log.Printf("CreateUser: building response: %v", err)
// 		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
// 		return
// 	}

// 	writeJSON(w, http.StatusCreated, resp)

// }

// func validateCreateUserRequest(req CreateUserRequest) error {
// 	if req.Username == "" {
// 		return errRequired("username")
// 	}
// 	if req.DisplayName == "" {
// 		return errRequired("display_name")
// 	}
// 	if req.PasswordHash == "" {
// 		return errRequired("password_hash")
// 	}

// 	uLength := utf8.RuneCountInString(req.Username)
// 	dLength := utf8.RuneCountInString(req.DisplayName)

// 	if uLength > maxUsernameLength {
// 		return fmt.Errorf("Username too long. Length: %d. Max length: %d", uLength, maxUsernameLength)
// 	}

// 	if dLength > maxDisplayNameLength {
// 		return fmt.Errorf("Display name too long. Length: %d. Max length: %d", dLength, maxDisplayNameLength)
// 	}

// 	return nil
// }

func (s *Server) GetUser(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")

	if err := validateGetUserRequest(id); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, err.Error())
		return
	}

	user, err := s.store.QueryUser(r.Context(), id)
	if err != nil {
		log.Printf("Get user: %v", err)
		writeError(w, err)
		return
	}

	userID, err := persistence.FromPgUUID(user.ID)
	if err != nil {
		log.Printf("GetUser: FromPgUUID: %v", err)
		writeError(w, err)
		return
	}

	if AuthenticateUserID(r) != userID {
		writeErrorMessage(w, http.StatusUnauthorized, "Unauthorized")
		return
	}

	resp, err := toUserResponse(user)
	if err != nil {
		log.Printf("CreateUser: building response: %v", err)
		writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
		return
	}

	writeJSON(w, http.StatusOK, resp)
}

func validateGetUserRequest(id string) error {
	if id == "" {
		return errRequired("id")
	}
	if utf8.RuneCountInString(id) != 36 {
		return errInvalid("id must be a valid user id")
	}

	return nil
}
