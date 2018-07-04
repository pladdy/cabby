package main

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

func withAcceptStix(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _ := splitAcceptHeader(r.Header.Get("Accept"))

		if contentType != stixContentType {
			unsupportedMediaType(w, fmt.Errorf("Invalid 'Accept' Header: %v", contentType))
			return
		}
		h(w, r)
	}
}

func withAcceptTaxii(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _ := splitAcceptHeader(r.Header.Get("Accept"))

		if contentType != taxiiContentType {
			unsupportedMediaType(w, fmt.Errorf("Invalid 'Accept' Header: %v", contentType))
			return
		}
		h(w, r)
	}
}

func withBasicAuth(h http.Handler, ts taxiiStorer) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		tu, _ := newTaxiiUser(ts, u, p)

		if !ok || !tu.valid() {
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}

		r = withTaxiiUser(r, tu)
		h.ServeHTTP(withHSTS(w), r)
	})
}

func withRequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userName).(string)
		if !ok {
			unauthorized(w, errors.New("Invalid user"))
		}

		milliSecondOfNanoSeconds := int64(1000000)

		start := time.Now().In(time.UTC)
		log.WithFields(log.Fields{
			"method":   r.Method,
			"start_ts": start.UnixNano() / milliSecondOfNanoSeconds,
			"url":      r.URL,
			"user":     user,
		}).Info("Request made to server")

		h(w, r)

		end := time.Now().In(time.UTC)
		elapsed := time.Since(start)
		log.WithFields(log.Fields{
			"elapsed_ts": float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
			"method":     r.Method,
			"end_ts":     end.UnixNano() / milliSecondOfNanoSeconds,
			"url":        r.URL,
			"user":       user,
		}).Info("Request made to server")
	}
}

/* helpers */
func splitAcceptHeader(h string) (string, string) {
	parts := strings.Split(h, ";")
	first := parts[0]

	var second string
	if len(parts) > 1 {
		second = parts[1]
	}

	return first, second
}
