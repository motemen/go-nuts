package httputil

import (
	"net/http"
)

type ErrorHandlerFunc func(http.ResponseWriter, *http.Request) error

func (f ErrorHandlerFunc) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	err := f(w, r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
