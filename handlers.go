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
	"time"

	log "github.com/sirupsen/logrus"
)

// per context docuentation, use a key type for context keys
type key int

const (
	userName        key = 0
	userCollections key = 1
	requestRange    key = 2
)

const (
	stixContentType20  = "application/vnd.oasis.stix+json; version=2.0"
	stixContentType    = "application/vnd.oasis.stix+json"
	taxiiContentType20 = "application/vnd.oasis.taxii+json; version=2.0"
	taxiiContentType   = "application/vnd.oasis.taxii+json"
)

type taxiiFilter struct {
	addedAfter   string
	collectionID string
	pagination   taxiiRange
	stixID       string
	stixTypes    []string
	version      string
}

func newTaxiiFilter(r *http.Request) (tf taxiiFilter) {
	tf.addedAfter = takeAddedAfter(r)
	tf.collectionID = takeCollectionID(r)
	tf.pagination = takeRequestRange(r)
	tf.stixID = takeStixID(r)
	tf.stixTypes = takeStixTypes(r)
	tf.version = takeVersion(r)
	return
}

type taxiiRange struct {
	first int64
	last  int64
	total int64
}

// newRange returns a Range given a string from the 'Range' HTTP header string
// the Range HTTP Header is specified by the request with the syntax 'items X-Y'
func newTaxiiRange(items string) (tr taxiiRange, err error) {
	tr = taxiiRange{first: -1, last: -1}

	if items == "" {
		return tr, err
	}

	itemDelimiter := "-"
	raw := strings.TrimSpace(items)
	tokens := strings.Split(raw, itemDelimiter)

	if len(tokens) == 2 {
		tr.first, err = strconv.ParseInt(tokens[0], 10, 64)
		tr.last, err = strconv.ParseInt(tokens[1], 10, 64)
		return tr, err
	}
	return tr, errors.New("Invalid range specified")
}

func (t *taxiiRange) Valid() bool {
	if t.first < 0 && t.last < 0 {
		return false
	}

	if t.first > t.last {
		return false
	}

	return true
}

func (t *taxiiRange) String() string {
	s := "items " +
		strconv.FormatInt(t.first, 10) +
		"-" +
		strconv.FormatInt(t.last, 10)

	if t.total > 0 {
		s += "/" + strconv.FormatInt(t.total, 10)
	}

	return s
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

func takeAddedAfter(r *http.Request) string {
	q := r.URL.Query()
	af := q["added_after"]

	if len(af) > 0 {
		t, err := time.Parse(time.RFC3339Nano, af[0])
		if err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}
	return ""
}

func takeCollectionAccess(r *http.Request) taxiiCollectionAccess {
	ctx := r.Context()

	// get collection access map from userCollections context
	ca, ok := ctx.Value(userCollections).(map[taxiiID]taxiiCollectionAccess)
	if !ok {
		return taxiiCollectionAccess{}
	}

	tid, err := newTaxiiID(takeCollectionID(r))
	if err != nil {
		return taxiiCollectionAccess{}
	}
	return ca[tid]
}

func takeCollectionID(r *http.Request) string {
	var collectionIndex = 3
	return getToken(r.URL.Path, collectionIndex)
}

func takeObjectID(r *http.Request) string {
	var objectIDIndex = 5
	return getToken(r.URL.Path, objectIDIndex)
}

func takeRequestRange(r *http.Request) taxiiRange {
	ctx := r.Context()

	tr, ok := ctx.Value(requestRange).(taxiiRange)
	if !ok {
		return taxiiRange{}
	}
	return tr
}

func takeStatusID(r *http.Request) string {
	var statusIndex = 3
	return getToken(r.URL.Path, statusIndex)
}

func takeStixID(r *http.Request) string {
	q := r.URL.Query()
	si := q["match[id]"]

	if len(si) > 0 {
		return si[0]
	}
	return ""
}

func takeStixTypes(r *http.Request) []string {
	q := r.URL.Query()
	st := q["match[type]"]

	if len(st) > 0 {
		return strings.Split(st[0], ",")
	}
	return []string{}
}

func takeVersion(r *http.Request) string {
	q := r.URL.Query()
	v := q["match[version]"]

	if len(v) > 0 {
		return v[0]
	}
	return ""
}

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

func withTaxiiRange(r *http.Request, tr taxiiRange) *http.Request {
	ctx := context.WithValue(r.Context(), requestRange, tr)
	return r.WithContext(ctx)
}

/* helpers */

func getToken(s string, i int) string {
	tokens := strings.Split(s, "/")

	if len(tokens) > i {
		return tokens[i]
	}
	return ""
}

func getAPIRoot(p string) string {
	var rootIndex = 1
	return getToken(p, rootIndex)
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

func writePartialContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusPartialContent)
	io.WriteString(w, content)
}
