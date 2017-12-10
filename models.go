package main

import (
	"encoding/json"
	"io/ioutil"
)

type apiRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

type config struct {
	Host       string
	Port       int
	SSLCert    string `json:"ssl_cert"`
	SSLKey     string `json:"ssl_key"`
	Discovery  discoveryResource
	APIRootMap map[string]apiRoot `json:"api_root_map"`
}

func (c *config) discoveryDefined() bool {
	if c.Discovery.Title == "" {
		return false
	}
	return true
}

// check if given api root is in the API Root Map of a config
func (c config) inAPIRootMap(ar string) (r bool) {
	for k := range c.APIRootMap {
		if ar == k {
			r = true
		}
	}
	return
}

// check if given api root is in the API Roots of a config
func (c config) inAPIRoots(ar string) (r bool) {
	for _, v := range c.Discovery.APIRoots {
		if ar == v {
			r = true
		}
	}
	return
}

// given a path to a config file parse it from json
func (c config) parse(file string) (pc config) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		warn.Panic(err)
	}

	if err = json.Unmarshal(b, &pc); err != nil {
		warn.Panic(err)
	}

	return
}

// validate that an api root has a same definition in apiRootMap
func (c config) validAPIRoot(a string) bool {
	return c.inAPIRoots(a) && c.inAPIRootMap(a)
}

type discoveryResource struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

type taxiiError struct {
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	ErrorID         string            `json:"error_id,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	HTTPStatus      int               `json:"http_status,string,omitempty"`
	ExternalDetails string            `json:"external_details,omitempty"`
	Details         map[string]string `json:"details,omitempty"`
}
