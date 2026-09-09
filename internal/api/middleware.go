package api

import (
	"log"
	"net/http"
	"time"
)

func withMiddleware(h http.Handler) http.Handler {
	return recoverPanic(logRequests(h))
}

// logRequests logs method, path, status, and duration for every request.
func logRequests(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(sw, r)
		log.Printf("%s %s %d %s", r.Method, r.URL.Path, sw.status, time.Since(start))
	})
}

// recoverPanic ensures a panic in any handler returns a 500 instead of crashing the whole server process
func recoverPanic(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if rec := recover(); rec != nil {
				log.Printf("panic handling %s %s: %v", r.Method, r.URL.Path, rec)
				writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
			}
		}()
		next.ServeHTTP(w, r)
	})
}

// statusWriter captures the status code so logRequests can report it
type statusWriter struct {
	http.ResponseWriter
	status int
}

func (sw *statusWriter) WriteHeader(status int) {
	sw.status = status
	sw.ResponseWriter.WriteHeader(status)
}
