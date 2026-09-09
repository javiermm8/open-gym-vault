package api

import "net/http"

// Routes table(using http.NewServeMux())
func (s *Server) routes() http.Handler {
	mux := http.NewServeMux()

	mux.HandleFunc("POST /sessions", s.CreateSession)
	mux.HandleFunc("POST /newUser", s.CreateUser)
	mux.HandleFunc("POST /newExercise", s.CreateExercise)

	return withMiddleware(mux)
}
