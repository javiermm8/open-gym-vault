package api

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
	"github.com/javiermm8/open-gym-vault/internal/persistence"
)

const cookieName = "opengymvault_token"

var publicOperations = map[string]bool{
	"LoginUser":    true,
	"RegisterUser": true,
	"LogoutUser":   true,
}

type ctxKeyUserID struct{}
type ctxKeyToken struct{}

func extractToken(r *http.Request) (string, bool) {
	if h := r.Header.Get("Authorization"); strings.HasPrefix(h, "Bearer ") {
		if token := strings.TrimPrefix(h, "Bearer "); len(token) == 43 {
			return token, true
		}
	}
	if cookie, err := r.Cookie(cookieName); err == nil && len(cookie.Value) == 43 {
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

func (s *Server) Authenticate(f gen.StrictHandlerFunc, operationID string) gen.StrictHandlerFunc {
	return func(ctx context.Context, w http.ResponseWriter, r *http.Request, request any) (any, error) {
		token, ok := extractToken(r)
		if ok {
			ctx = context.WithValue(ctx, ctxKeyToken{}, token)
		}
		if publicOperations[operationID] {
			return f(ctx, w, r, request)
		}

		if !ok {
			writeErrorMessage(w, http.StatusUnauthorized, "authentication required")
			return nil, nil
		}

		userID, err := s.store.ExtendTokenExpiry(ctx, token)
		if err != nil {
			if errors.Is(err, persistence.ErrTokenInvalidOrExpired) {
				writeErrorMessage(w, http.StatusUnauthorized, "invalid or expired token")
				return nil, nil
			}
			return nil, err
		}

		return f(context.WithValue(ctx, ctxKeyUserID{}, userID), w, r, request)
	}
}

func AuthenticateUserID(ctx context.Context) string {
	id, ok := ctx.Value(ctxKeyUserID{}).(string)
	if !ok {
		panic("AuthenticateUserID called on a request not wrapped with Authenticate")
	}
	return id
}

func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(ctxKeyToken{}).(string)
	return token, ok
}
