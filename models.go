package main

import (
	"encoding/json"
	"io/ioutil"
)

/* Config */

type Config struct {
	Host      string
	Port      int
	SSLCert   string
	SSLKey    string
	Discovery DiscoveryResource
}

type CabbyConfig struct {
	Config
}

type RethinkConfig struct {
	Config
}

func (c Config) parse(file string) (config Config) {
	info.Println("Parsing config file", file)

	b, err := ioutil.ReadFile(file)
	if err != nil {
		warn.Panic(err)
	}

	err = json.Unmarshal(b, &config)
	if err != nil {
		warn.Panic(err)
	}

	//debug.Println("Config:", string(b))
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
