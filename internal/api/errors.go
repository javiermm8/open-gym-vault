package api

import (
	"errors"
	"net/http"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// From https://www.postgresql.org/docs/current/errcodes-appendix.html
const (
	sqlStateCheckViolation      = "23514"
	sqlStateForeignKeyViolation = "23503"
	sqlStateUniqueViolation     = "23505"
)

// Handle db errors and give them an appropiate http status code
func writeError(w http.ResponseWriter, err error) {
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case sqlStateCheckViolation:
			writeErrorMessage(w, http.StatusBadRequest, "invalid data: "+pgErr.Message)
			return
		case sqlStateForeignKeyViolation:
			writeErrorMessage(w, http.StatusBadRequest, "references a resource that does not exist")
			return
		case sqlStateUniqueViolation:
			writeErrorMessage(w, http.StatusConflict, "resource already exists")
			return
		}
	}

	if errors.Is(err, pgx.ErrNoRows) {
		writeErrorMessage(w, http.StatusNotFound, "not found")
		return
	}

	writeErrorMessage(w, http.StatusInternalServerError, "internal server error")
}
