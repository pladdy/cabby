package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
)

/* Config */

type Config struct {
	Host    string
	Port    int
	SSLCert string
	SSLKey  string
}

type CabbyConfig struct {
	Config
}

type RethinkConfig struct {
	Config
}

func (c Config) parse(file string) (config Config) {
	configFile, err := os.Open(file)
	if err != nil {
		warn.Panic(err)
	}

	decoder := json.NewDecoder(configFile)
	err = decoder.Decode(&config)
	if err != nil {
		warn.Panic(err)
	}

	return
}

/* Discovery */

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
		warn.Panic(err)
	}
	return b
}

/* Error */

type Error struct {
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	ErrorId         string            `json:"error_id,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	HTTPStatus      int               `json:"http_status,string,omitempty"`
	ExternalDetails string            `json:"external_details,omitempty"`
	Details         map[string]string `json:"details,omitempty"`
}

func (e *Error) Message() string {
	b, err := json.Marshal(e)
	if err != nil {
		warn.Panic(err)
	}

	return string(b)
}
