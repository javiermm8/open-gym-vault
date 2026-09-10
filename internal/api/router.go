package api

import "net/http"

// Routes table(using http.NewServeMux())
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	// POSTs
	mux.HandleFunc("POST /new_session", s.CreateSession)
	mux.HandleFunc("POST /new_user", s.CreateUser)
	mux.HandleFunc("POST /new_exercise", s.CreateExercise)

	// GETs
	mux.HandleFunc("GET /session/{id}", s.GetSession)
	mux.HandleFunc("GET /user/{id}", s.GetUser)

	return withMiddleware(mux)
}
