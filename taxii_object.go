package main

import (
	"errors"
	"net/http"
)

func handleTaxiiObjects(ts taxiiStorer, maxContentLength int64) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodPost:
			if contentTooLarge(r.ContentLength, maxContentLength) {
				requestTooLarge(w, r.ContentLength, maxContentLength)
				return
			}
			handlePostTaxiiObjects(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
		}
	})
}

func contentTooLarge(r, m int64) bool {
	if r > m {
		return true
	}
	return false
}

func handlePostTaxiiObjects(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	s, err := newTaxiiStatus()
	if err != nil {
		resourceNotFound(w, errors.New("Unable to process status"))
	}

	writeContent(w, taxiiContentType, resourceToJSON(s))
}
