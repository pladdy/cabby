package main

import (
	"errors"
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

// TODO: make this function require data to be posted and that data is parsed as JSON to a struct
func handlePostTaxiiCollection(w http.ResponseWriter, r *http.Request) {
	err := r.ParseForm()
	if err != nil {
		badRequest(w, err)
		return
	}

	tc, err := newTaxiiCollection(r.Form.Get("id"))
	if err != nil {
		badRequest(w, err)
		return
	}
	tc.Title = r.Form.Get("title")
	tc.Description = r.Form.Get("description")

	user, ok := r.Context().Value(userName).(string)
	if !ok {
		badRequest(w, errors.New("Invalid user specified"))
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

// creating a collection is a two step process: create the collection then create the association of the collection
// to the api root
func (tc taxiiCollection) create(user, apiRoot string) error {
	var err error

	parts := []struct {
		name string
		args []interface{}
	}{
		{"taxiiCollection", []interface{}{tc.ID.String(), tc.Title, tc.Description, strings.Join(tc.MediaTypes, ",")}},
		{"taxiiCollectionAPIRoot", []interface{}{tc.ID.String(), apiRoot}},
		{"taxiiUserCollection", []interface{}{user, tc.ID.String(), true, true}},
	}

	for _, p := range parts {
		err := createTaxiiCollectionPart(p.name, p.args)
		if err != nil {
			return err
		}
	}

	return err
}

func createTaxiiCollectionPart(part string, args []interface{}) error {
	t, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	query, err := t.parse("create", part)
	if err != nil {
		return err
	}

	toWrite := make(chan interface{}, minBuffer)
	errs := make(chan error, minBuffer)

	go t.write(query, toWrite, errs)
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

func (tc *taxiiCollection) read(u string) error {
	collection := *tc

	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	query, err := ts.parse("read", "taxiiCollection")
	if err != nil {
		return err
	}

	result, err := ts.read(query, "taxiiCollection", []interface{}{u, tc.ID.String()})
	if err != nil {
		return err
	}

	tcs := result.(taxiiCollections)
	collection = firstTaxiiCollection(tcs)
	*tc = collection

	return err
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

	query, err := ts.parse("read", "taxiiCollections")
	if err != nil {
		return err
	}

	args := []interface{}{u}
	result, err := ts.read(query, "taxiiCollections", args)
	if err != nil {
		return err
	}

	collections = result.(taxiiCollections)
	*tcs = collections
	return err
}
