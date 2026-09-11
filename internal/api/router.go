package api

import "net/http"

// Routes table(using http.NewServeMux()) (gonna be my todos now)
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// TODOs:
	// - update profile
	// - Rate limits
	// - Stats
	// - Session templates
	// - Docs/openapi
	// - Better log error mgsgs

	// AUTH
	mux.HandleFunc("POST /auth/register", s.Register)
	mux.HandleFunc("POST /auth/login", s.Login)
	mux.HandleFunc("POST /auth/logout", s.Logout)

	// POSTs
	mux.HandleFunc("POST /new_session", s.RequireAuth(s.CreateSession))
	// mux.HandleFunc("POST /new_user", s.CreateUser) // New user is now register, add update profile instead
	mux.HandleFunc("POST /new_exercise", s.RequireAuth(s.CreateExercise))

	// GETs by ID
	mux.HandleFunc("GET /session/{id}", s.RequireAuth(s.GetSession))
	mux.HandleFunc("GET /user/{id}", s.RequireAuth(s.GetUser))
	mux.HandleFunc("GET /exercise/{id}", s.RequireAuth(s.GetExercise))

	// GETs lists
	mux.HandleFunc("GET /exercises", s.RequireAuth(s.ListExercises))
	mux.HandleFunc("GET /sessions", s.RequireAuth(s.ListSessions))

	return withMiddleware(mux)
}
