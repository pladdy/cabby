package main

import (
	"errors"
	"net/http"
)

/* handler */

func handleTaxiiDiscovery(ts taxiiStorer) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		info.Println("Discovery resource requested")
		defer recoverFromPanic(w)

		td := taxiiDiscovery{}

		err := td.read(ts)
		td.Default = insertPort(td.Default)

		if err != nil {
			badRequest(w, err)
			return
		}

		if td.Title == "" {
			resourceNotFound(w, errors.New("Discovery not defined"))
		} else {
			writeContent(w, taxiiContentType, resourceToJSON(td))
		}
	})
}

/* model */

type taxiiDiscovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

func (td *taxiiDiscovery) create(ts taxiiStorer) error {
	err := createResource(ts, "taxiiDiscovery", []interface{}{td.Title, td.Description, td.Contact, td.Default})
	return err
}

func (td *taxiiDiscovery) read(ts taxiiStorer) error {
	discovery := *td

	result, err := ts.read("taxiiDiscovery", []interface{}{})
	if err != nil {
		return err
	}
	discovery = result.(taxiiDiscovery)

	*td = discovery
	return err
}
