package main

import (
	"context"
	"errors"
	"net/http"

	log "github.com/sirupsen/logrus"
)

// per context docuentation, use a key type for context keys
type key int

const (
	userName         key = 0
	userCollections  key = 1
	maxContentLength key = 2
)

const (
	sixMonthsOfSeconds = "63072000"
)

func takeCollectionAccess(r *http.Request) taxiiCollectionAccess {
	ctx := r.Context()

	// get collection access map from userCollections context
	ca, ok := ctx.Value(userCollections).(map[taxiiID]taxiiCollectionAccess)
	if !ok {
		return taxiiCollectionAccess{}
	}

	tid, err := newTaxiiID(getCollectionID(r.URL.Path))
	if err != nil {
		return taxiiCollectionAccess{}
	}
	return ca[tid]
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

		r = withTaxiiUser(tu, r)
		h.ServeHTTP(withHSTS(w), r)
	})
}

func withHSTS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
	return w
}

func withTaxiiUser(tu taxiiUser, r *http.Request) *http.Request {
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
