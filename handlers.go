package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
)

/* auth functions */

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !validate(user, pass) {
			warn.Println("Invalid user/pass combination")
			w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
			error := Error{Title: "Unauthorized", HTTPStatus: http.StatusUnauthorized}
			http.Error(w, error.Message(), http.StatusUnauthorized)
			return
		}

		info.Println("Basic Auth validated")
		h(w, r)
	}
}

func validate(u, p string) bool {
	if u == "pladdy" && p == "pants" {
		return true
	}
	return false
}

/* resource handlers */

func strictTransportSecurity() (key, value string) {
	return "Strict-Transport-Security", "max-age=" + strconv.Itoa(SixMonthsOfSeconds) + "; includeSubDomains"
}

func resourceNotFound(w http.ResponseWriter) {
	if r := recover(); r != nil {
		http.Error(w, "Resource not found", http.StatusNotFound)
	}
}

/* discovery */

func handleDiscovery(w http.ResponseWriter, r *http.Request) {
	defer resourceNotFound(w)
	info.Println("Discovery resource requested")

	config := Config{}.parse(ConfigPath)
	verifyDiscoveryDefined(config.Discovery)

	b, err := json.Marshal(config.Discovery)
	if err != nil {
		warn.Panic("Can't serve Discovery:", err)
	}

	w.Header().Set("Content-Type", TAXIIContentType)
	info.Println("Responding with a Discovery resource")
	io.WriteString(w, string(b))
}

func verifyDiscoveryDefined(d DiscoveryResource) {
	if d.Title == "" {
		warn.Panic("Discovery Resource not defined")
	}
}

/* register configured API Roots */

func handleAPIRoot(w http.ResponseWriter, r *http.Request) {
	defer resourceNotFound(w)

	u := removePort(r.URL)
	info.Println("API Root requested for", u)

	config := Config{}.parse(ConfigPath)
	fmt.Println(config)

	if !config.validAPIRoot(u) {
		warn.Panic("API Root ", u, " not defined in config file")
	}

	b, err := json.Marshal(config.APIRootMap[u])
	if err != nil {
		warn.Panic("Can't serve ", u, " error:", err)
	}

	w.Header().Set("Content-Type", TAXIIContentType)
	io.WriteString(w, string(b))
}

func registerAPIRoots(h *http.ServeMux) {
	config := Config{}.parse(ConfigPath)

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

func removePort(u *url.URL) string {
	noPort := u.Scheme + "://" + u.Hostname() + u.Path
	return noPort
}

/* wrappers for handlers */

func HSTS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(strictTransportSecurity())
		h.ServeHTTP(w, r)
	})
}
