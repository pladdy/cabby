package main

import (
	"encoding/json"
	"io/ioutil"
)

/* API root */

type APIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

/* Config */

type Config struct {
	Host       string
	Port       int
	SSLCert    string `json:"ssl_cert"`
	SSLKey     string `json:"ssl_key"`
	Discovery  DiscoveryResource
	APIRootMap map[string]APIRoot `json:"api_root_map"`
}

// given a path to a config file parse it from json
func (c Config) parse(file string) (config Config) {
	info.Println("Parsing config file", file)

	b, err := ioutil.ReadFile(file)
	if err != nil {
		warn.Panic(err)
	}

	if err = json.Unmarshal(b, &config); err != nil {
		warn.Panic(err)
	}

	return
}

// check if given api root is in the API Root Map of a config
func (c Config) inAPIRootMap(ar string) (r bool) {
	for k, _ := range c.APIRootMap {
		if ar == k {
			r = true
		}
	}
	return
}

// check if given api root is in the API Roots of a config
func (c Config) inAPIRoots(ar string) (r bool) {
	for _, v := range c.Discovery.APIRoots {
		if ar == v {
			r = true
		}
	}
	return
}

// validate that an api root has a same definition in APIRootMap
func (c Config) validAPIRoot(a string) bool {
	return c.inAPIRoots(a) && c.inAPIRootMap(a)
}

/* Discovery */

type DiscoveryResource struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
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
