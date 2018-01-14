package main

import (
	"encoding/json"
	"errors"
	"fmt"
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
			logWarn.Println("Invalid user/pass combination")
			unauthorized(w)
			return
		}
		logInfo.Println("Basic Auth validated for", user)
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

func badRequest(w http.ResponseWriter, err error) {
	logError.Println(err)
	errString := fmt.Sprintf("%v", err)

	te := taxiiError{Title: "Bad Request", Description: errString, HTTPStatus: http.StatusBadRequest}
	http.Error(w, resourceToJSON(te), http.StatusBadRequest)
}

func resourceNotFound(w http.ResponseWriter) {
	te := taxiiError{Title: "Resource not found", HTTPStatus: http.StatusNotFound}
	http.Error(w, resourceToJSON(te), http.StatusNotFound)
}

func recoverFromPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		resourceNotFound(w)
	}
}

/* api root */

func handleTaxiiAPIRoot(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w)

	u := urlWithNoPort(r.URL)
	logInfo.Println("API Root requested for", u)

	config := cabbyConfig{}.parse(configPath)

	if !config.validAPIRoot(u) {
		logWarn.Panic("API Root ", u, " not defined in config file")
	}

	w.Header().Set("Content-Type", taxiiContentType)
	io.WriteString(w, resourceToJSON(config.APIRootMap[u]))
}

/* collections */

func handleTaxiiCollection(w http.ResponseWriter, r *http.Request) {
	defer recoverFromPanic(w)

	switch r.Method {
	case "POST":
		postTaxiiCollection(w, r)
	default:
		badRequest(w, errors.New("HTTP Method "+r.Method+" Unrecognized"))
		return
	}
}

func postTaxiiCollection(w http.ResponseWriter, r *http.Request) {
	var err error

	// any error, send badRequest
	defer func(err *error) {
		if err != nil {
			err := fmt.Errorf("%v", err)
			badRequest(w, err)
		}
	}(&err)

	err = r.ParseForm()
	if err != nil {
		return
	}

	tc, err := newTaxiiCollection(r.Form.Get("id"))
	if err != nil {
		return
	}
	tc.Title = r.Form.Get("title")
	tc.Description = r.Form.Get("description")

	err = tc.create(cabbyConfig{}.parse(configPath))
	if err != nil {
		return
	}

	w.Header().Set("Content-Type", taxiiContentType)
	io.WriteString(w, resourceToJSON(tc))
}

/* discovery */

func handleTaxiiDiscovery(w http.ResponseWriter, r *http.Request) {
	logInfo.Println("Discovery resource requested")
	defer recoverFromPanic(w)

	config := cabbyConfig{}.parse(configPath)
	if config.discoveryDefined() == false {
		logWarn.Panic("Discovery Resource not defined")
	}

	w.Header().Set("Content-Type", taxiiContentType)
	io.WriteString(w, resourceToJSON(config.Discovery))
}

/* catch undefined route */

func handleUndefinedRequest(w http.ResponseWriter, r *http.Request) {
	logWarn.Printf("Undefined request: %v\n", r.URL)
	resourceNotFound(w)
}

/* helpers */

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		logWarn.Panicf("Can't convert %v to JSON, error: %v", v, err)
	}
	return string(b)
}

func urlWithNoPort(u *url.URL) string {
	c := cabbyConfig{}.parse(configPath)
	var noPort string

	if u.Host == "" {
		noPort = "https://" + c.Host + u.Path
	} else {
		noPort = u.Scheme + "://" + u.Hostname() + u.Path
	}
	return noPort
}
