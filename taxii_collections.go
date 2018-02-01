package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"

	uuid "github.com/satori/go.uuid"
)

/* handlers */

func handleTaxiiCollections(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w)

	switch r.Method {
	case "GET":
		handleGetTaxiiCollections(w, r)
	case "POST":
		handlePostTaxiiCollection(w, r)
	default:
		badRequest(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
		return
	}
}

func handleGetTaxiiCollections(w http.ResponseWriter, r *http.Request) {
	info.Println("Request for GET Collection:", r.URL)

	id := lastURLPathToken(r.URL.Path)

	user, ok := r.Context().Value(userName).(string)
	if !ok {
		badRequest(w, errors.New("Invalid user specified"))
		return
	}

	if id == "collections" {
		readTaxiiCollections(w, user)
	} else {
		readTaxiiCollection(w, id, user)
	}
}

func taxiiCollectionFromBytes(b []byte) (taxiiCollection, error) {
	var tc taxiiCollection

	err := json.Unmarshal(b, &tc)
	if err != nil {
		return tc, fmt.Errorf("invalid data to POST, error: %v", err)
	}

	err = tc.ensureID()
	if err != nil {
		return tc, err
	}
	tc.CanRead = true
	tc.CanWrite = true

	return tc, err
}

func handlePostTaxiiCollection(w http.ResponseWriter, r *http.Request) {
	info.Println("Request to POST Collection:", r.URL)

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		badRequest(w, err)
		return
	}
	defer r.Body.Close()

	tc, err := taxiiCollectionFromBytes(body)
	if err != nil {
		fail.Println("from bytes failed")
		badRequest(w, err)
		return
	}

	user, ok := r.Context().Value(userName).(string)
	if !ok {
		badRequest(w, errors.New("No user specified"))
		return
	}

	err = tc.create(user, apiRoot(r.URL.Path))
	if err != nil {
		badRequest(w, err)
		return
	}

	writeContent(w, taxiiContentType, resourceToJSON(tc))
}

func readTaxiiCollection(w http.ResponseWriter, id, user string) {
	tc, err := newTaxiiCollection(id)
	if err != nil {
		badRequest(w, err)
		return
	}

	err = tc.read(user)
	if err != nil {
		badRequest(w, err)
		return
	}

	if tc.ID.String() != id {
		resourceNotFound(w, errors.New("Invalid Collection"))
	} else {
		writeContent(w, taxiiContentType, resourceToJSON(tc))
	}
}

func readTaxiiCollections(w http.ResponseWriter, user string) {
	tcs := taxiiCollections{}
	err := tcs.read(user)
	if err != nil {
		badRequest(w, err)
		return
	}

	if len(tcs.Collections) == 0 {
		resourceNotFound(w, errors.New("No collections available for this user"))
	} else {
		writeContent(w, taxiiContentType, resourceToJSON(tcs))
	}
}

/* models */

type taxiiID struct {
	uuid.UUID
}

func newTaxiiID(arg ...string) (taxiiID, error) {
	if len(arg) > 0 && len(arg[0]) > 0 {
		id, err := uuid.FromString(arg[0])
		return taxiiID{id}, err
	}

	id, err := uuid.NewV4()
	return taxiiID{id}, err
}

func (ti taxiiID) isEmpty() bool {
	empty := taxiiID{}
	if ti == empty {
		return true
	}
	return false
}

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
		tc.ID, err = newTaxiiID(id[0])
	}

	tc.MediaTypes = []string{taxiiContentType}
	return tc, err
}

// creating a collection is a multi-step process, multiple "parts" have to be created as part of the associations
func (tc taxiiCollection) create(user, apiRoot string) error {
	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	parts := []struct {
		name string
		args []interface{}
	}{
		{"taxiiCollection", []interface{}{tc.ID.String(), tc.Title, tc.Description, strings.Join(tc.MediaTypes, ",")}},
		{"taxiiCollectionAPIRoot", []interface{}{tc.ID.String(), apiRoot}},
		{"taxiiUserCollection", []interface{}{user, tc.ID.String(), true, true}},
	}

	for _, p := range parts {
		err := createTaxiiCollectionPart(ts, p.name, p.args)
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

func (tc *taxiiCollection) read(u string) error {
	collection := *tc

	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	tq, err := ts.parse("read", "taxiiCollection")
	if err != nil {
		return err
	}

	result, err := ts.read(tq, []interface{}{u, tc.ID.String()})
	if err != nil {
		return err
	}

	tcs := result.(taxiiCollections)
	collection = firstTaxiiCollection(tcs)
	*tc = collection

	return err
}

/* taxiiCollection helpers */

func createTaxiiCollectionPart(ts taxiiStorer, part string, args []interface{}) error {
	tq, err := ts.parse("create", part)
	if err != nil {
		return err
	}

	toWrite := make(chan interface{}, minBuffer)
	errs := make(chan error, minBuffer)

	go ts.write(tq, toWrite, errs)
	toWrite <- args
	close(toWrite)

	for e := range errs {
		err = e
	}

	return err
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

func (tcs *taxiiCollections) read(u string) error {
	collections := *tcs

	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	tq, err := ts.parse("read", "taxiiCollections")
	if err != nil {
		return err
	}

	args := []interface{}{u}
	result, err := ts.read(tq, args)
	if err != nil {
		return err
	}

	collections = result.(taxiiCollections)
	*tcs = collections
	return err
}
