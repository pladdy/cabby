package main

import (
	"context"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

const (
	sixMonthsOfSeconds = "63072000"
)

// decorate a handler with basic authentication
func withBasicAuth(h http.Handler, ts taxiiStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		tu, validated := validateUser(ts, u, p)

		if !ok || !validated {
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}

		r = withTaxiiUser(r, tu)
		h.ServeHTTP(withHSTS(w), r)
	})
}

func withHSTS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
	return w
}

func withTaxiiUser(r *http.Request, tu taxiiUser) *http.Request {
	ctx := context.WithValue(context.Background(), userName, tu.Email)
	ctx = context.WithValue(ctx, userCollections, tu.CollectionAccess)
	return r.WithContext(ctx)
}

func validateUser(ts taxiiStorer, u, p string) (taxiiUser, bool) {
	tu, err := newTaxiiUser(ts, u, p)
	if err != nil {
		log.Error(err)
		return tu, false
	}

	return tu, true
}
