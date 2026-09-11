package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"unicode/utf8"

	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

const minPasswordLength = 8

func (s *Server) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Username == "" || req.DisplayName == "" || req.Password == "" {
		writeErrorMessage(w, http.StatusBadRequest, "username, display_name and password are required")
		return
	}
	if utf8.RuneCountInString(req.Password) < minPasswordLength {
		writeErrorMessage(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	}

	user, err := s.store.Register(r.Context(), req.Username, req.DisplayName, req.Password)
	if err != nil {
		log.Printf("Register: %v", err)
		writeError(w, err)
		return
	}

	writeJSON(w, http.StatusCreated, map[string]string{
		"id":           user.ID,
		"username":     user.Username,
		"display_name": user.DisplayName,
	})
}

func (s *Server) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErrorMessage(w, http.StatusBadRequest, "invalid JSON body")
		return
	}
	if req.Username == "" || req.Password == "" {
		writeErrorMessage(w, http.StatusBadRequest, "username and password are required")
		return
	}
	rawToken, expiresAt, user, err := s.store.Login(r.Context(), req.Username, req.Password)
	if err != nil {
		if errors.Is(err, persistence.ErrInvalidCredentials) {
			writeErrorMessage(w, http.StatusUnauthorized, "invalid username or password")
			return
		}
		log.Printf("Login: %v", err)
		writeError(w, err)
		return
	}

	setAuthCookie(w, rawToken, expiresAt)

	writeJSON(w, http.StatusOK, authResponse{
		Token:     rawToken,
		ExpiresAt: expiresAt,
		UserID:    user.ID,
	})
}

func (s *Server) Logout(w http.ResponseWriter, r *http.Request) {
	if token, ok := extractToken(r); ok {
		if err := s.store.Logout(r.Context(), token); err != nil {
			log.Printf("Logout: %v", err)
			writeErrorMessage(w, http.StatusInternalServerError, "failed to log out")
			return
		}
	} else {
		writeErrorMessage(w, http.StatusBadRequest, "authentication required")
		return
	}
	clearAuthCookie(w)
	writeJSON(w, http.StatusOK, map[string]string{"status": "logged out"})
}
