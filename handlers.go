package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/url"
)

const (
	sixMonthsOfSeconds = "63072000"
	stixContentType    = "application/vnd.oasis.stix+json; version=2.0"
	taxiiContentType   = "application/vnd.oasis.taxii+json; version=2.0"
)

func hsts(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
		h.ServeHTTP(w, r)
	})
}

/* auth functions */

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !validated(user, pass) {
			warn.Println("Invalid user/pass combination")
			unauthorized(w)
			return
		}
		info.Println("Basic Auth validated")
		h(w, r)
	}
}

func validated(u, p string) bool {
	if u == "pladdy" && p == "pants" {
		return true
	}
	return false
}

/* http status functions */

func unauthorized(w http.ResponseWriter) {
	te := taxiiError{Title: "Unauthorized", Description: "Invalid user/password combination", HTTPStatus: http.StatusUnauthorized}
	w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
	http.Error(w, resourceToJSON(te), http.StatusUnauthorized)
}

func resourceNotFound(w http.ResponseWriter) {
	if r := recover(); r != nil {
		te := taxiiError{Title: "Resource not found", HTTPStatus: http.StatusNotFound}
		http.Error(w, resourceToJSON(te), http.StatusNotFound)
	}
}

/* handlers */

func handleDiscovery(w http.ResponseWriter, r *http.Request) {
	info.Println("Discovery resource requested")
	defer resourceNotFound(w)

	config := config{}.parse(configPath)
	if config.discoveryDefined() == false {
		warn.Panic("Discovery Resource not defined")
	}

	w.Header().Set("Content-Type", taxiiContentType)
	io.WriteString(w, resourceToJSON(config.Discovery))
}

/* register configured API Roots */

func handleAPIRoot(w http.ResponseWriter, r *http.Request) {
	defer resourceNotFound(w)

	u := urlWithNoPort(r.URL)
	info.Println("API Root requested for", u)

	config := config{}.parse(configPath)

	if !config.validAPIRoot(u) {
		warn.Panic("API Root ", u, " not defined in config file")
	}

	w.Header().Set("Content-Type", taxiiContentType)
	io.WriteString(w, resourceToJSON(config.APIRootMap[u]))
}

func registerAPIRoots(h *http.ServeMux) {
	config := config{}.parse(configPath)

	for _, apiRoot := range config.Discovery.APIRoots {
		if config.validAPIRoot(apiRoot) {
			u, err := url.Parse(apiRoot)
			if err != nil {
				warn.Panic(err)
			}

			info.Println("Registering API handler for", u)
			h.HandleFunc(u.Path, basicAuth(handleAPIRoot))
		}
	}
}

/* helpers */

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		warn.Panic("Can't convert %v to JSON, error: ", v, err)
	}
	return string(b)
}

func urlWithNoPort(u *url.URL) string {
	c := config{}.parse(configPath)
	var noPort string

	if u.Host == "" {
		noPort = "https://" + c.Host + u.Path
	} else {
		noPort = u.Scheme + "://" + u.Hostname() + u.Path
	}
	return noPort
}
