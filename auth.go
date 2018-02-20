package main

import (
	"context"
	"errors"
	"net/http"
)

// per context docuentation, use a key type for context keys
type key int

const (
	sixMonthsOfSeconds     = "63072000"
	userName           key = 0
	userCollections    key = 1
)

func addTaxiiUserToRequest(tu taxiiUser, r *http.Request) *http.Request {
	ctx := context.WithValue(context.Background(), userName, tu.Email)
	ctx = context.WithValue(ctx, userCollections, tu.CollectionAccess)
	return r.WithContext(ctx)
}

// decorate a handler with basic authentication
func withBasicAuth(ts taxiiStorer, h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		tu, validated := validateUser(ts, u, p)

		if !ok || !validated {
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}
		info.Println("Basic Auth validated for", u)

		r = addTaxiiUserToRequest(tu, r)
		h.ServeHTTP(withHSTS(w), r)
	})
}

func withHSTS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
	return w
}

func validateUser(ts taxiiStorer, u, p string) (taxiiUser, bool) {
	tu, err := newTaxiiUser(ts, u, p)
	if err != nil {
		fail.Println(err)
		return tu, false
	}

	return tu, true
}
