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
