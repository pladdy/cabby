package main

import (
	"errors"
	"net/http"
)

/* handler */

func handleTaxiiDiscovery(w http.ResponseWriter, r *http.Request) {
	info.Println("Discovery resource requested")
	defer recoverFromPanic(w)

	td := taxiiDiscovery{}

	err := td.read()
	if err != nil {
		badRequest(w, err)
		return
	}

	if td.Title == "" {
		resourceNotFound(w, errors.New("Discovery not defined"))
	} else {
		writeContent(w, taxiiContentType, resourceToJSON(td))
	}
}

/* model */

type taxiiDiscovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

func (td *taxiiDiscovery) create() error {
	err := createResource("taxiiDiscovery", []interface{}{td.Title, td.Description, td.Contact, td.Default})
	return err
}

func (td *taxiiDiscovery) read() error {
	discovery := *td

	result, err := readResource("taxiiDiscovery", []interface{}{})
	if err != nil {
		return err
	}
	discovery = result.(taxiiDiscovery)

	*td = discovery
	return err
}
