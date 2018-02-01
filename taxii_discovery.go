package main

import (
	"net/http"
	"net/url"
	"strconv"
)

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

func insertPort(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		fail.Panic(err)
	}
	return u.Scheme + "://" + u.Host + ":" + strconv.Itoa(config.Port) + u.Path
}

func (td *taxiiDiscovery) read() {
	discovery := *td

	if config.discoveryDefined() == false {
		fail.Panic("Discovery Resource not defined in config")
	}

	discovery = config.Discovery
	discovery.Default = insertPort(discovery.Default)

	var apiRootsWithPort []string
	for _, apiRoot := range discovery.APIRoots {
		apiRootsWithPort = append(apiRootsWithPort, insertPort(apiRoot))
	}
	discovery.APIRoots = apiRootsWithPort

	*td = discovery
}
