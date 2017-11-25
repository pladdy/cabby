package main

import (
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

const (
	TAXIIContentType      = "application/vnd.oasis.taxii+json; version=2.0"
	DiscoveryResourceFile = "data/discovery.json"
)

type DiscoveryResource struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Contact     string   `json:"contact"`
	Default     string   `json:"default"`
	APIRoots    []string `json:"api_roots"`
}

func parseDiscoveryResource(resource string) []byte {
	b, err := ioutil.ReadFile(resource)
	if err != nil {
		log.Panic(err)
	}
	return b
}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !validate(user, pass) {
			w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h(w, r)
	}
}

func validate(u, p string) bool {
	if u == "pladdy" && p == "pants" {
		return true
	}
	return false
}

func handleDiscovery(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "Resource not found", http.StatusNotFound)
		}
	}()

	b := parseDiscoveryResource(DiscoveryResourceFile)
	w.Header().Set("Content-Type", TAXIIContentType)
	io.WriteString(w, string(b))
}

func main() {
	http.HandleFunc("/taxii", basicAuth(handleDiscovery))
	http.ListenAndServe(":1234", nil)
}
