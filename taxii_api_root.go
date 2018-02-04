package main

import (
	"errors"
	"net/http"
	"regexp"
	"strings"
)

/* handler */

func trimSlahes(s string) string {
	re := regexp.MustCompile("^/")
	s = re.ReplaceAllString(s, "")

	re = regexp.MustCompile("/$")
	s = re.ReplaceAllString(s, "")

	parts := strings.Split(s, "/")
	return strings.Join(parts, "/")
}

func handleTaxiiAPIRoot(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w)
	info.Println("API Root requested for", r.URL)

	ta := taxiiAPIRoot{}

	info.Println("reading path:", trimSlahes(r.URL.Path))

	err := ta.read(trimSlahes(r.URL.Path))
	if err != nil {
		badRequest(w, err)
		return
	}

	if ta.Title == "" {
		resourceNotFound(w, errors.New("API Root not defined"))
	} else {
		writeContent(w, taxiiContentType, resourceToJSON(ta))
	}
}

/* model */

type taxiiAPIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

func (ta *taxiiAPIRoot) create(path string) error {
	id, err := newTaxiiID()
	if err != nil {
		return err
	}

	err = createResource("taxiiAPIRoot",
		[]interface{}{id, path, ta.Title, ta.Description, strings.Join(ta.Versions, ","), ta.MaxContentLength})
	return err
}

func (ta *taxiiAPIRoot) read(path string) error {
	apiRoot := *ta

	result, err := readResource("taxiiAPIRoot", []interface{}{path})
	if err != nil {
		return err
	}
	apiRoot = result.(taxiiAPIRoot)

	*ta = apiRoot
	return err
}
