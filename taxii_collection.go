package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

/* handlers */

func handleTaxiiCollections(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodGet:
			handleGetTaxiiCollections(ts, w, r)
		case http.MethodPost:
			handlePostTaxiiCollection(ts, w, r)
		default:
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}
	})
}

func handleGetTaxiiCollections(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	user, ok := r.Context().Value(userName).(string)
	if !ok {
		unauthorized(w, errors.New("Invalid user specified"))
		return
	}

	tr, err := newTaxiiRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err)
		return
	}
	r = withTaxiiRange(r, tr)

	result, err := getTaxiiCollections(ts, r, user)
	if err != nil {
		resourceNotFound(w, err)
		return
	}

	if tr.Valid() {
		tr.total = result.items
		w.Header().Set("Content-Range", tr.String())
		writePartialContent(w, taxiiContentType, resourceToJSON(result.data))
	} else {
		writeContent(w, taxiiContentType, resourceToJSON(result.data))
	}
}

func getTaxiiCollections(ts taxiiStorer, r *http.Request, user string) (taxiiResult, error) {
	if lastURLPathToken(r.URL.Path) == "collections" {
		return readTaxiiCollections(ts, r, user)
	}
	return readTaxiiCollection(ts, r, user)
}

func taxiiCollectionFromBytes(b []byte) (taxiiCollection, error) {
	var tc taxiiCollection

	err := json.Unmarshal(b, &tc)
	if err != nil {
		return tc, fmt.Errorf("Invalid data to POST, error: %v", err)
	}

	err = tc.ensureID()
	if err != nil {
		return tc, err
	}
	tc.CanRead = true
	tc.CanWrite = true

	return tc, err
}

func handlePostTaxiiCollection(ts taxiiStorer, w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w, err)
		return
	}
	defer r.Body.Close()

	tc, err := taxiiCollectionFromBytes(body)
	if err != nil {
		badRequest(w, err)
		return
	}

	user, ok := r.Context().Value(userName).(string)
	if !ok {
		unauthorized(w, errors.New("No user specified"))
		return
	}

	err = tc.create(ts, user, getAPIRoot(r.URL.Path))
	if err != nil {
		internalServerError(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tc))
}

func readTaxiiCollection(ts taxiiStorer, r *http.Request, user string) (taxiiResult, error) {
	var result taxiiResult
	id := lastURLPathToken(r.URL.Path)

	tc, err := newTaxiiCollection(id)
	if err != nil {
		return result, err
	}

	result, err = tc.read(ts, user)
	if err != nil {
		return result, err
	}

	// if the read returns no data
	if tc.ID.String() != id {
		return result, errors.New("Invalid collection")
	}
	return result, err
}

func readTaxiiCollections(ts taxiiStorer, r *http.Request, user string) (taxiiResult, error) {
	tf := newTaxiiFilter(r)
	tcs := taxiiCollections{}

	result, err := tcs.read(ts, user, tf)
	if err != nil {
		return result, err
	}

	if len(tcs.Collections) == 0 {
		return result, errors.New("No collections available for this user")
	}
	return result, err
}

/* models */

type taxiiCollection struct {
	ID          taxiiID  `json:"id"`
	CanRead     bool     `json:"can_read"`
	CanWrite    bool     `json:"can_write"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"`
}

func newTaxiiCollection(id ...string) (taxiiCollection, error) {
	var err error
	tc := taxiiCollection{}

	// try to set id to a taxii id unless it's 'collections'
	if id[0] != "collections" {
		tc.ID, err = taxiiIDFromString(id[0])
	}

	tc.MediaTypes = []string{taxiiContentType}
	return tc, err
}

// creating a collection is a multi-step process, multiple "parts" have to be created as part of the associations
func (tc *taxiiCollection) create(ts taxiiStorer, user, apiRoot string) error {
	var err error

	parts := []struct {
		resource string
		args     []interface{}
	}{
		{"taxiiCollection", []interface{}{tc.ID.String(), apiRoot, tc.Title, tc.Description, strings.Join(tc.MediaTypes, ",")}},
		{"taxiiUserCollection", []interface{}{user, tc.ID.String(), true, true}},
	}

	for _, p := range parts {
		err = createResource(ts, p.resource, p.args)
		if err != nil {
			return err
		}
	}

	return err
}

func (tc *taxiiCollection) ensureID() error {
	var err error
	if tc.ID.isEmpty() {
		tc.ID, err = newTaxiiID()
	}
	return err
}

func (tc *taxiiCollection) read(ts taxiiStorer, u string) (taxiiResult, error) {
	collection := *tc

	result, err := ts.read("taxiiCollection", []interface{}{u, tc.ID.String()})
	if err != nil {
		return result, err
	}

	tcs := result.data.(taxiiCollections)
	collection = firstTaxiiCollection(tcs)
	*tc = collection

	return result, err
}

func firstTaxiiCollection(tcs taxiiCollections) taxiiCollection {
	if len(tcs.Collections) > 0 {
		return tcs.Collections[0]
	}
	return taxiiCollection{}
}

type taxiiCollections struct {
	Collections []taxiiCollection `json:"collections"`
}

func (tcs *taxiiCollections) read(ts taxiiStorer, u string, tf taxiiFilter) (taxiiResult, error) {
	collections := *tcs

	result, err := ts.read("taxiiCollections", []interface{}{u}, tf)
	if err != nil {
		return result, err
	}

	collections = result.data.(taxiiCollections)
	*tcs = collections
	return result, err
}
