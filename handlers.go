package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// per context docuentation, use a key type for context keys
type key int

const (
	sixMonthsOfSeconds = "63072000"

	stixContentType20  = "application/vnd.oasis.stix+json; version=2.0"
	stixContentType    = "application/vnd.oasis.stix+json"
	taxiiContentType20 = "application/vnd.oasis.taxii+json; version=2.0"
	taxiiContentType   = "application/vnd.oasis.taxii+json"

	userName        key = 0
	userCollections key = 1
)

/* auth functions */

func addTaxiiUserToRequest(tu taxiiUser, r *http.Request) *http.Request {
	ctx := context.WithValue(context.Background(), userName, tu.Email)
	ctx = context.WithValue(ctx, userCollections, tu.CollectionAccess)
	return r.WithContext(ctx)
}

// decorate a handler with basic authentication
func basicAuth(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		u, p, ok := r.BasicAuth()
		tu, validated := validateUser(u, p)

		if !ok || !validated {
			unauthorized(w, errors.New("Invalid user/pass combination"))
			return
		}
		info.Println("Basic Auth validated for", u)

		r = addTaxiiUserToRequest(tu, r)
		h.ServeHTTP(hsts(w), r)
	})
}

func hsts(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
	return w
}

func splitAcceptHeader(h string) (string, string) {
	parts := strings.Split(h, ";")
	first := parts[0]

	var second string
	if len(parts) > 1 {
		second = parts[1]
	}

	return first, second
}

func requireAcceptTaxii(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		contentType, _ := splitAcceptHeader(r.Header.Get("Accept"))

		if contentType != taxiiContentType {
			unsupportedMediaType(w, fmt.Errorf("Invalid 'Accept' Header: %v", contentType))
			return
		}
		h(w, r)
	}
}

func validateUser(u, p string) (taxiiUser, bool) {
	tu, err := newTaxiiUser(u, p)
	if err != nil {
		fail.Println(err)
		return tu, false
	}

	return tu, true
}

/* http status functions */

func errorStatus(w http.ResponseWriter, title string, err error, status int) {
	fail.Println(err)
	errString := fmt.Sprintf("%v", err)

	te := taxiiError{Title: title, Description: errString, HTTPStatus: status}
	w.Header().Set("Content-Type", taxiiContentType)
	http.Error(w, resourceToJSON(te), status)
}

func badRequest(w http.ResponseWriter, err error) {
	errorStatus(w, "Bad Request", err, http.StatusBadRequest)
}

func methodNotAllowed(w http.ResponseWriter, err error) {
	errorStatus(w, "Method Not Allowed", err, http.StatusMethodNotAllowed)
}

func resourceNotFound(w http.ResponseWriter, err error) {
	errorStatus(w, "Resource not found", err, http.StatusNotFound)
}

func unauthorized(w http.ResponseWriter, err error) {
	w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
	errorStatus(w, "Unauthorized", err, http.StatusUnauthorized)
}

func unsupportedMediaType(w http.ResponseWriter, err error) {
	errorStatus(w, "Unsupported Media Type", err, http.StatusUnsupportedMediaType)
}

/* catch undefined route */

func handleUndefinedRequest(w http.ResponseWriter, r *http.Request) {
	resourceNotFound(w, fmt.Errorf("Undefined request: %v", r.URL))
}

func recoverFromPanic(w http.ResponseWriter) {
	if r := recover(); r != nil {
		resourceNotFound(w, errors.New("Resource not found"))
	}
}

/* helpers */

func apiRoot(u string) string {
	var rootIndex = 1
	tokens := strings.Split(u, "/")
	return tokens[rootIndex]
}

func lastURLPathToken(u string) string {
	u = strings.TrimSuffix(u, "/")
	tokens := strings.Split(u, "/")
	length := len(tokens)
	return tokens[length-1]
}

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		warn.Panicf("Can't convert %v to JSON, error: %v", v, err)
	}
	return string(b)
}

func urlWithNoPort(u *url.URL) string {
	var noPort string

	if u.Host == "" {
		noPort = "https://" + config.Host + u.Path
	} else {
		noPort = u.Scheme + "://" + u.Hostname() + u.Path
	}
	return noPort
}

func writeContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	io.WriteString(w, content)
}
