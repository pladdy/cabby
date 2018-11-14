package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

func handleUndefinedRoute(w http.ResponseWriter, r *http.Request) {
	resourceNotFound(w, fmt.Errorf("Invalid path: %v", r.URL))
}

// RequestHandler interface for handling requests
type RequestHandler interface {
	Get(w http.ResponseWriter, r *http.Request)
	Post(w http.ResponseWriter, r *http.Request)
}

// RouteRequest takes a RequestHandler and routes requests to its methods
func RouteRequest(h RequestHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodGet:
			h.Get(w, r)
		case http.MethodPost:
			h.Post(w, r)
		case http.MethodHead:
			// for HEAD requests send to GET, it will omit response
			h.Get(w, r)
		default:
			w.Header().Set("Allow", "Get, Head, Post")
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
			return
		}
	})
}

func verifySupportedMimeType(w http.ResponseWriter, r *http.Request, mh, mv string) bool {
	mimeType, _ := splitMimeType(r.Header.Get(mh))

	if mimeType != mv {
		unsupportedMediaType(w, fmt.Errorf("Invalid '%v' Header: %v", mh, mimeType))
		return false
	}
	return true
}

// WithMimeType decorates a handle with content type check
func WithMimeType(h http.HandlerFunc, mh, mv string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if !verifySupportedMimeType(w, r, mh, mv) {
			return
		}
		h(w, r)
	}
}

func withBasicAuth(h http.Handler, us cabby.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		r = withTransactionID(r)

		user, err := us.User(r.Context(), u, p)
		if err != nil {
			internalServerError(w, err)
			return
		}

		if !ok || !user.Defined() {
			log.WithFields(log.Fields{"user": u}).Warn("User authentication failed!	")
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}

		ucs, err := us.UserCollections(r.Context(), u)
		if err != nil {
			internalServerError(w, err)
			return
		}
		user.CollectionAccessList = ucs.CollectionAccessList

		log.WithFields(log.Fields{"user": u}).Info("User authenticated")
		h.ServeHTTP(withHSTS(w), r.WithContext(cabby.WithUser(r.Context(), user)))
	})
}

func withLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		milliSecondOfNanoSeconds := int64(1000000)

		start := time.Now().In(time.UTC)
		log.WithFields(log.Fields{
			"method":                 r.Method,
			"request_content_length": r.ContentLength,
			"start_ms":               start.UnixNano() / milliSecondOfNanoSeconds,
			"transaction_id":         cabby.TakeTransactionID(r.Context()),
			"url":                    r.URL.String(),
			"user":                   cabby.TakeUser(r.Context()).Email,
		}).Info("Request received")

		h.ServeHTTP(w, r)

		end := time.Now().In(time.UTC)
		elapsed := time.Since(start)

		log.WithFields(log.Fields{
			"bytes":          cabby.TakeBytes(r.Context()),
			"elapsed_ms":     float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
			"end_ms":         end.UnixNano() / milliSecondOfNanoSeconds,
			"method":         r.Method,
			"transaction_id": cabby.TakeTransactionID(r.Context()),
			"url":            r.URL.String(),
			"user":           cabby.TakeUser(r.Context()).Email,
		}).Info("Request served")
	})
}

/* helpers */

func splitMimeType(h string) (string, string) {
	parts := strings.Split(h, ";")
	first := parts[0]

	var second string
	if len(parts) > 1 {
		second = parts[1]
	}

	return first, second
}
