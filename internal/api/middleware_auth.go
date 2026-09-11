package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

const cookieName = "opengymvault_token"

func extractToken(r *http.Request) (string, bool) {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		return strings.TrimPrefix(h, "Bearer "), true
	}
	if cookie, err := r.Cookie(cookieName); err == nil && cookie.Value != "" {
		return cookie.Value, true
	}
	return "", false
}

func setAuthCookie(w http.ResponseWriter, token string, expiresAt time.Time) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    token,
		Expires:  expiresAt,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func clearAuthCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    "",
		Expires:  time.Unix(0, 0),
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteLaxMode,
		Path:     "/",
	})
}

func (s *Server) RequireAuth(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		token, ok := extractToken(r)
		if !ok {
			writeErrorMessage(w, http.StatusUnauthorized, "authentication required")
			return
		}

		userID, err := s.store.ExtendTokenExpiry(r.Context(), token)
		if err != nil {
			if errors.Is(err, persistence.ErrTokenInvalidOrExpired) {
				writeErrorMessage(w, http.StatusUnauthorized, "invalid or expired token")
				return
			}
			writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
			return
		}

		ctx := context.WithValue(r.Context(), "userID", userID)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}

func AuthenticateUserID(r *http.Request) uuid.UUID {
	id, ok := r.Context().Value("userID").(uuid.UUID)
	if !ok {
		panic("AuthenticateUserID called on a request not wrapped with RequireAuth")
	}
	return id
}
