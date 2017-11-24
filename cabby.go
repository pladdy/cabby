package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

const TAXIIContentType = "application/vnd.oasis.taxii+json; version=2.0"

type DiscoveryResource struct {
	Title       string   `json:"title"`
	Description string   `json:"description"`
	Contact     string   `json:"contact"`
	Default     string   `json:"default"`
	APIRoots    []string `json:"api_roots"`
}

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !validate(user, pass) {
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
	w.Header().Set("Content-Type", TAXIIContentType)

	resource := DiscoveryResource{
		"Test Discovery",
		"This is a test discovery resource",
		"pladdy",
		"https://test.com/api1",
		[]string{"https://test.com/api2", "https://test.com/api3"}}

	b, err := json.Marshal(resource)
	if err == nil {
		fmt.Fprintf(w, string(b))
	}
}

func main() {
	http.HandleFunc("/taxii", basicAuth(handleDiscovery))
	http.ListenAndServe(":1234", nil)
}
