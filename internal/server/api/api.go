// Package api provides HTTP handlers for the application.
package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/davidsbond/keeper/internal/server/token"
)

type (
	// The Error type represents an error as returned by API endpoints.
	Error struct {
		// The error message.
		Message string `json:"message"`
		// The error code, should correspond to the HTTP status code used in the response.
		Code int `json:"code"`
	}

	// The Validatable interface describes types that can be validated.
	Validatable interface {
		// Validate should check the underlying type's validity and return an error if invalid.
		Validate() error
	}
)

func (e Error) Error() string {
	return fmt.Sprintf("%s (%d)", e.Message, e.Code)
}

func decode[T Validatable](r io.Reader) (T, error) {
	var out T
	if err := json.NewDecoder(r).Decode(&out); err != nil {
		return out, err
	}

	return out, out.Validate()
}

func writeErrorf(w http.ResponseWriter, code int, message string, args ...any) {
	writeError(w, code, fmt.Sprintf(message, args...))
}

func writeError(w http.ResponseWriter, code int, message string) {
	write(w, code, Error{
		Message: message,
		Code:    code,
	})
}

func write(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(body); err != nil {
		// TODO(davidsbond): something
		fmt.Println(err)
	}
}

func requireToken(next http.HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !token.FromContext(r.Context()).Valid() {
			writeError(w, http.StatusUnauthorized, http.StatusText(http.StatusUnauthorized))
			return
		}

		next.ServeHTTP(w, r)
	})
}
