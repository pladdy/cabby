package main

import "net/http"

/* handler */

func handleTaxiiDiscovery(w http.ResponseWriter, r *http.Request) {
	info.Println("Discovery resource requested")
	defer recoverFromPanic(w)

	td := taxiiDiscovery{}
	td.read()

	writeContent(w, taxiiContentType, resourceToJSON(td))
}

/* model */

type taxiiDiscovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

func (td *taxiiDiscovery) read() {
	discovery := *td

	if config.discoveryDefined() == false {
		fail.Panic("Discovery Resource not defined in config")
	}

	discovery = config.Discovery
	*td = discovery
}
