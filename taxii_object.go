package main

import (
	"errors"
	"net/http"
)

func handleTaxiiObjects(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodPost:
			handlePostTaxiiObjects(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
		}
	})
}

func handlePostTaxiiObjects(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	s, err := newTaxiiStatus()
	if err != nil {
		resourceNotFound(w, errors.New("Unable to process status"))
	}

	writeContent(w, stixContentType, resourceToJSON(s))
}
