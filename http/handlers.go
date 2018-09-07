package http

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gofrs/uuid"
	cabby "github.com/pladdy/cabby2"
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
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
			return
		}
	})
}

// WithAcceptType decorates a handle with content type check
func WithAcceptType(h http.HandlerFunc, typeToCheck string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _ := splitAcceptHeader(r.Header.Get("Accept"))

		if contentType != typeToCheck {
			unsupportedMediaType(w, fmt.Errorf("Invalid 'Accept' Header: %v", contentType))
			return
		}
		h(w, r)
	}
}

func withBasicAuth(h http.Handler, us cabby.UserService) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()

		user, err := us.User(r.Context(), p)
		if err != nil {
			internalServerError(w, err)
			return
		}

		if !ok || !user.Defined() {
			log.WithFields(log.Fields{"user": u}).Warn("User authentication failed!	")
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}

		ucs, err := us.UserCollections(r.Context())
		if err != nil {
			internalServerError(w, err)
			return
		}
		user.CollectionAccessList = ucs.CollectionAccessList

		log.WithFields(log.Fields{"user": u}).Info("User authenticated")
		h.ServeHTTP(withHSTS(w), r.WithContext(cabby.WithUser(r.Context(), user)))
	})
}

func withRequestLogging(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		milliSecondOfNanoSeconds := int64(1000000)
		transactionID := uuid.Must(uuid.NewV4())
		r.WithContext(cabby.WithTransactionID(r.Context(), transactionID))

		start := time.Now().In(time.UTC)
		log.WithFields(log.Fields{
			"method":         r.Method,
			"start_ms":       start.UnixNano() / milliSecondOfNanoSeconds,
			"transaction_id": transactionID,
			"url":            r.URL.String(),
		}).Info("Request received")

		h.ServeHTTP(w, r)

		end := time.Now().In(time.UTC)
		elapsed := time.Since(start)

		log.WithFields(log.Fields{
			"elapsed_ts":     float64(elapsed.Nanoseconds()) / float64(milliSecondOfNanoSeconds),
			"method":         r.Method,
			"end_ms":         end.UnixNano() / milliSecondOfNanoSeconds,
			"transaction_id": transactionID,
			"url":            r.URL.String(),
		}).Info("Request served")
	})
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
