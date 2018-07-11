package main

import (
	"context"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

// per context docuentation, use a key type for context keys
type key int

const (
	userName           key = 0
	userCollections    key = 1
	requestRange       key = 2
	canAdmin           key = 3
	jsonContentType        = "application/json"
	sixMonthsOfSeconds     = "63072000"
	stixContentType20      = "application/vnd.oasis.stix+json; version=2.0"
	stixContentType        = "application/vnd.oasis.stix+json"
	taxiiContentType20     = "application/vnd.oasis.taxii+json; version=2.0"
	taxiiContentType       = "application/vnd.oasis.taxii+json"
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

/* request helpers */

func requestMethodIsGet(r *http.Request) bool {
	if r.Method == http.MethodGet {
		return true
	}
	return false
}

func takeAddedAfter(r *http.Request) string {
	af := r.URL.Query()["added_after"]

	if len(af) > 0 {
		t, err := time.Parse(time.RFC3339Nano, af[0])
		if err == nil {
			return t.Format(time.RFC3339Nano)
		}
	}
	return ""
}

func takeCanAdmin(r *http.Request) bool {
	ca, ok := r.Context().Value(canAdmin).(bool)
	if !ok {
		return false
	}
	return ca
}

func takeCollectionAccess(r *http.Request) taxiiCollectionAccess {
	// get collection access map from userCollections context
	ca, ok := r.Context().Value(userCollections).(map[taxiiID]taxiiCollectionAccess)
	if !ok {
		return taxiiCollectionAccess{}
	}

	tid, err := taxiiIDFromString(takeCollectionID(r))
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
	si := r.URL.Query()["match[id]"]

	if len(si) > 0 {
		return si[0]
	}
	return ""
}

func takeStixTypes(r *http.Request) []string {
	st := r.URL.Query()["match[type]"]

	if len(st) > 0 {
		return strings.Split(st[0], ",")
	}
	return []string{}
}

func takeUser(r *http.Request) string {
	user, ok := r.Context().Value(userName).(string)
	if !ok {
		return ""
	}
	return user
}

func takeVersion(r *http.Request) string {
	v := r.URL.Query()["match[version]"]

	if len(v) > 0 {
		return v[0]
	}
	return ""
}

func userExists(r *http.Request) bool {
	_, ok := r.Context().Value(userName).(string)
	if !ok {
		return false
	}
	return true
}

func withTaxiiRange(r *http.Request, tr taxiiRange) *http.Request {
	ctx := context.WithValue(r.Context(), requestRange, tr)
	return r.WithContext(ctx)
}

func withTaxiiUser(r *http.Request, tu taxiiUser) *http.Request {
	ctx := context.WithValue(context.Background(), userName, tu.Email)
	ctx = context.WithValue(ctx, canAdmin, tu.CanAdmin)
	ctx = context.WithValue(ctx, userCollections, tu.CollectionAccess)
	return r.WithContext(ctx)
}

func validateUser(ts taxiiStorer, u, p string) (taxiiUser, bool) {
	tu, err := newTaxiiUser(ts, u, p)
	if err != nil {
		log.Error(err)
		return tu, false
	}

	if !tu.valid() {
		log.WithFields(log.Fields{"user": u}).Error("Invalid user")
		return tu, false
	}

	return tu, true
}
