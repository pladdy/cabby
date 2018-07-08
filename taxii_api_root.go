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
			resourceNotFound(w, err)
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
	Path             string   `json:"path,omitempty"`
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

func (ta *taxiiAPIRoot) create(ts taxiiStorer) error {
	id, err := taxiiIDUsingString(ta.Path)
	if err != nil {
		return err
	}

	return createResource(ts, "taxiiAPIRoot",
		[]interface{}{id, ta.Path, ta.Title, ta.Description, strings.Join(ta.Versions, ","), ta.MaxContentLength})
}

func (ta *taxiiAPIRoot) delete(ts taxiiStorer) error {
	return ts.delete("taxiiAPIRoot", []interface{}{ta.Path})
}

func (ta *taxiiAPIRoot) read(ts taxiiStorer, path string) error {
	apiRoot := *ta

	result, err := ts.read("taxiiAPIRoot", []interface{}{path})
	if err != nil {
		return err
	}
	apiRoot = result.data.(taxiiAPIRoot)

	*ta = apiRoot
	return err
}

func (ta *taxiiAPIRoot) update(ts taxiiStorer) error {
	return ts.update("taxiiAPIRoot",
		[]interface{}{ta.Title, ta.Description, strings.Join(ta.Versions, ","), ta.MaxContentLength, ta.Path})
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
	roots = result.data.(taxiiAPIRoots)

	*ta = roots
	return err
}
