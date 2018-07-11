package main

import (
	"errors"
	"net/http"
	"strings"
)

/* handlers */

func handleTaxiiCollections(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !requestMethodIsGet(r) {
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
			return
		}

		if !userExists(r) {
			unauthorized(w, errors.New("No user specified"))
			return
		}

		tr, err := newTaxiiRange(r.Header.Get("Range"))
		if err != nil {
			rangeNotSatisfiable(w, err)
			return
		}
		r = withTaxiiRange(r, tr)

		result, err := getTaxiiCollections(ts, r, takeUser(r))
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
	})
}

func getTaxiiCollections(ts taxiiStorer, r *http.Request, user string) (taxiiResult, error) {
	if lastURLPathToken(r.URL.Path) == "collections" {
		return readTaxiiCollections(ts, r, user)
	}
	return readTaxiiCollection(ts, r, user)
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
	APIRootPath string   `json:"api_root_path,omitempty"`
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

func (tc *taxiiCollection) delete(ts taxiiStorer) error {
	return ts.delete("taxiiCollection", []interface{}{tc.ID.String()})
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

func (tc *taxiiCollection) update(ts taxiiStorer) error {
	return ts.update("taxiiCollection",
		[]interface{}{tc.ID.String(), tc.APIRootPath, tc.Title, tc.Description, tc.ID.String()})
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
