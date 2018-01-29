package main

import (
	"net/http"
)

/* handler */

func handleTaxiiAPIRoot(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w)

	u := urlWithNoPort(r.URL)
	info.Println("API Root requested for", u)

	ta := taxiiAPIRoot{}
	ta.read(u)

	writeContent(w, taxiiContentType, resourceToJSON(ta))
}

/* model */

type taxiiAPIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

func (ta *taxiiAPIRoot) read(u string) {
	apiRoot := *ta

	if !config.validAPIRoot(u) {
		warn.Panic("API Root ", u, " not defined in config file")
	}

	apiRoot = config.APIRootMap[u]
	*ta = apiRoot
}
