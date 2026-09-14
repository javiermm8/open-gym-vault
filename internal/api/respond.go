package api

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
)

type loginResponse struct {
	gen.LoginUser200JSONResponse
	token     string
	expiresAt time.Time
}

type logoutResponse struct {
	gen.LogoutUser204Response
}

type changePasswordResponse struct {
	gen.ChangePassword204Response
}

func (resp loginResponse) VisitLoginUserResponse(w http.ResponseWriter) error {
	setAuthCookie(w, resp.token, resp.expiresAt)
	return resp.LoginUser200JSONResponse.VisitLoginUserResponse(w)
}

func (resp logoutResponse) VisitLogoutUserResponse(w http.ResponseWriter) error {
	clearAuthCookie(w)
	return resp.LogoutUser204Response.VisitLogoutUserResponse(w)
}

func (resp changePasswordResponse) VisitChangePasswordResponse(w http.ResponseWriter) error {
	clearAuthCookie(w)
	return resp.ChangePassword204Response.VisitChangePasswordResponse(w)
}

// writeJSON encodes v as JSON with the given status code
func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if v == nil {
		return
	}
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("error encoding JSON response: %v", err)
	}
}

type errorResponse struct {
	Error string `json:"error"`
}

func writeErrorMessage(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, errorResponse{Error: message})
}
