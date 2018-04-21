package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
)

// per context docuentation, use a key type for context keys
type key int

const (
	userName         key = 0
	userCollections  key = 1
	maxContentLength key = 2
	requestRange     key = 3
)

const (
	stixContentType20  = "application/vnd.oasis.stix+json; version=2.0"
	stixContentType    = "application/vnd.oasis.stix+json"
	taxiiContentType20 = "application/vnd.oasis.taxii+json; version=2.0"
	taxiiContentType   = "application/vnd.oasis.taxii+json"
)

type taxiiRange struct {
	first int64
	last  int64
}

// newRange returns a Range given a string from the 'Range' HTTP header string
// the Range HTTP Header is specified by the request with the syntax 'items X-Y'
func newTaxiiRange(items string) (hr taxiiRange, err error) {
	if items == "" {
		return hr, err
	}

	itemDelimiter := "-"
	raw := strings.TrimSpace(items)
	tokens := strings.Split(raw, itemDelimiter)

	if len(tokens) == 2 {
		hr.first, err = strconv.ParseInt(tokens[0], 10, 64)
		hr.last, err = strconv.ParseInt(tokens[1], 10, 64)
		return hr, err
	}
	return hr, errors.New("Invalid range specified")
}

func splitAcceptHeader(h string) (string, string) {
	parts := strings.Split(h, ";")
	first := parts[0]

	var second string
	if len(parts) > 1 {
		second = parts[1]
	}

	return first, second
}

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

func takeRequestRange(r *http.Request) taxiiRange {
	ctx := r.Context()

	tr, ok := ctx.Value(requestRange).(taxiiRange)
	if !ok {
		return taxiiRange{}
	}
	return tr
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

func withRequestLogging(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, ok := r.Context().Value(userName).(string)
		if !ok {
			unauthorized(w, errors.New("Invalid user"))
		}

		log.WithFields(log.Fields{
			"url":    r.URL,
			"method": r.Method,
			"user":   user,
		}).Info("Request made to server")

		h(w, r)
	}
}

func withTaxiiRange(r *http.Request, tr taxiiRange) *http.Request {
	ctx := context.WithValue(r.Context(), requestRange, tr)
	return r.WithContext(ctx)
}

/* http status functions */

func errorStatus(w http.ResponseWriter, title string, err error, status int) {
	errString := fmt.Sprintf("%v", err)

	te := taxiiError{Title: title, Description: errString, HTTPStatus: status}

	log.WithFields(log.Fields{
		"error":       err,
		"title":       title,
		"http status": status,
	}).Error("Returning error in response")

	w.Header().Set("Content-Type", taxiiContentType)
	http.Error(w, resourceToJSON(te), status)
}

func badRequest(w http.ResponseWriter, err error) {
	errorStatus(w, "Bad Request", err, http.StatusBadRequest)
}

func methodNotAllowed(w http.ResponseWriter, err error) {
	errorStatus(w, "Method Not Allowed", err, http.StatusMethodNotAllowed)
}

func resourceNotFound(w http.ResponseWriter, err error) {
	errorStatus(w, "Resource not found", err, http.StatusNotFound)
}

func requestTooLarge(w http.ResponseWriter, rc, mc int64) {
	err := fmt.Errorf("content length is %v, content length can't be bigger than %v", rc, mc)
	errorStatus(w, "Request too large", err, http.StatusRequestEntityTooLarge)
}

func rangeNotSatisfiable(w http.ResponseWriter, err error) {
	errorStatus(w, "Requested ange cannot be satisfied", err, http.StatusRequestedRangeNotSatisfiable)
}

func unauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
	errorStatus(w, "Unauthorized", err, http.StatusUnauthorized)
}

func unsupportedMediaType(w http.ResponseWriter, err error) {
	errorStatus(w, "Unsupported Media Type", err, http.StatusUnsupportedMediaType)
}

/* catch undefined route */

func handleUndefinedRequest(w http.ResponseWriter, r *http.Request) {
	resourceNotFound(w, fmt.Errorf("Undefined request: %v", r.URL))
}

func recoverFromPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		log.Error("Panic!")
		resourceNotFound(w, errors.New("Resource not found"))
	}
}

/* helpers */

func getToken(s string, i int) string {
	tokens := strings.Split(s, "/")

	if len(tokens) >= i {
		return tokens[i]
	}
	return ""
}

func getAPIRoot(p string) string {
	var rootIndex = 1
	return getToken(p, rootIndex)
}

func getCollectionID(p string) string {
	var collectionIndex = 3
	return getToken(p, collectionIndex)
}

func getStixID(p string) string {
	var stixIDIndex = 5
	return getToken(p, stixIDIndex)
}

func lastURLPathToken(u string) string {
	u = strings.TrimSuffix(u, "/")
	tokens := strings.Split(u, "/")
	return tokens[len(tokens)-1]
}

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.WithFields(log.Fields{
			"value": v,
			"error": err,
		}).Panic("Can't convert to JSON")
	}
	return string(b)
}

func writeContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	io.WriteString(w, content)
}
