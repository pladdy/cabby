package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

/* handler */

func trimSlashes(s string) string {
	re := regexp.MustCompile("^/")
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile("/$")
	s = re.ReplaceAllString(s, "")

	parts := strings.Split(s, "/")
	return strings.Join(parts, "/")
}

func handleTaxiiAPIRoot(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		ta := taxiiAPIRoot{}
		err := ta.read(ts, trimSlashes(r.URL.Path))
		if err != nil {
			badRequest(w, err)
			return
		}

		if ta.Title == "" {
			resourceNotFound(w, errors.New("API Root not defined"))
		} else {
			writeContent(w, taxiiContentType, resourceToJSON(ta))
		}
	})
}

/* models */

type taxiiAPIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

func (ta *taxiiAPIRoot) create(ts taxiiStorer, path string) error {
	id, err := newTaxiiID()
	if err != nil {
		return err
	}

	err = createResource(ts, "taxiiAPIRoot",
		[]interface{}{id, path, ta.Title, ta.Description, strings.Join(ta.Versions, ","), ta.MaxContentLength})
	return err
}

func (ta *taxiiAPIRoot) read(ts taxiiStorer, path string) error {
	apiRoot := *ta

	result, err := ts.read("taxiiAPIRoot", []interface{}{path})
	if err != nil {
		return err
	}
	apiRoot = result.(taxiiAPIRoot)

	*ta = apiRoot
	return err
}

type taxiiAPIRoots struct {
	RootPaths []string
}

func (ta *taxiiAPIRoots) read(ts taxiiStorer) error {
	roots := *ta

	result, err := ts.read("taxiiAPIRoots", []interface{}{})
	if err != nil {
		return err
	}
	roots = result.(taxiiAPIRoots)

	*ta = roots
	return err
}
