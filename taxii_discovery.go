package main

import (
	"errors"
	"net/http"
	"net/url"
	"strconv"

	log "github.com/sirupsen/logrus"
)

/* handler */

func handleTaxiiDiscovery(ts taxiiStorer, port int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		td := taxiiDiscovery{}
		err := td.read(ts)
		if err != nil {
			resourceNotFound(w, err)
			return
		}

		td.Default = insertPort(td.Default, port)

		for i := 0; i < len(td.APIRoots); i++ {
			td.APIRoots[i] = swapPath(td.Default, td.APIRoots[i]) + "/"
		}

		if td.Title == "" {
			resourceNotFound(w, errors.New("Discovery not defined"))
		} else {
			writeContent(w, taxiiContentType, resourceToJSON(td))
		}
	})
}

func urlTokens(u string) map[string]string {
	tokens, err := url.Parse(u)
	if err != nil {
		log.Panic("Can't parse url")
	}
	return map[string]string{"scheme": tokens.Scheme, "host": tokens.Host, "path": tokens.Path}
}

func insertPort(u string, port int) string {
	tokens := urlTokens(u)
	return tokens["scheme"] + "://" + tokens["host"] + ":" + strconv.Itoa(port) + tokens["path"]
}

func swapPath(u, p string) string {
	tokens := urlTokens(u)
	return tokens["scheme"] + "://" + tokens["host"] + "/" + p
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
	return createResource(ts, "taxiiDiscovery", []interface{}{td.Title, td.Description, td.Contact, td.Default})
}

func (td *taxiiDiscovery) delete(ts taxiiStorer) error {
	return ts.delete("taxiiDiscovery", []interface{}{})
}

func (td *taxiiDiscovery) read(ts taxiiStorer) error {
	discovery := *td

	result, err := ts.read("taxiiDiscovery", []interface{}{})
	if err != nil {
		return err
	}
	discovery = result.data.(taxiiDiscovery)

	*td = discovery
	return err
}

func (td *taxiiDiscovery) update(ts taxiiStorer) error {
	return ts.update("taxiiDiscovery", []interface{}{td.Title, td.Description, td.Contact, td.Default})
}
