package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
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

type Config struct {
	Host string
	Port int
}

func parseConfig(file string) (config Config) {
	configFile, err := os.Open(file)
	if err != nil {
		log.Panic(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		log.Panic(err)
	}

	return
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
	config := parseConfig("config.json")
	port := strconv.Itoa(config.Port)
	log.Println("Serving on port " + port)

	http.HandleFunc("/taxii", basicAuth(handleDiscovery))
	http.ListenAndServe(":"+port, nil)
}
