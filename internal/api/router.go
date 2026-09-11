package api

import "net/http"

// Routes table(using http.NewServeMux())
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// AUTH
	mux.HandleFunc("POST /auth/register", s.Register)
	mux.HandleFunc("POST /auth/login", s.Login)
	mux.HandleFunc("POST /auth/logout", s.Logout)

	// POSTs
	mux.HandleFunc("POST /new_session", s.CreateSession)
	mux.HandleFunc("POST /new_user", s.CreateUser)
	mux.HandleFunc("POST /new_exercise", s.CreateExercise)

	// GETs
	mux.HandleFunc("GET /session/{id}", s.RequireAuth(s.GetSession))
	mux.HandleFunc("GET /user/{id}", s.RequireAuth(s.GetUser))

	return withMiddleware(mux)
}
