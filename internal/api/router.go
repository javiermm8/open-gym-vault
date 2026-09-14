package api

import (
	"net/http"

	"github.com/javiermm8/open-gym-vault/internal/api/gen"
)

func (s *Server) routes() http.Handler {
	strict := gen.NewStrictHandlerWithOptions(s,
		[]gen.StrictMiddlewareFunc{s.Authenticate},
		gen.StrictHTTPServerOptions{
			RequestErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeErrorMessage(w, http.StatusBadRequest, err.Error())
			},
			ResponseErrorHandlerFunc: func(w http.ResponseWriter, r *http.Request, err error) {
				writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
			},
		},
	)
	return s.withMiddleware(gen.Handler(strict))
}
