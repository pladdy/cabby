package main

import (
	"encoding/json"
	"io/ioutil"
)

var config cabbyConfig

type cabbyConfig struct {
	Host      string
	Port      int
	SSLCert   string            `json:"ssl_cert"`
	SSLKey    string            `json:"ssl_key"`
	DataStore map[string]string `json:"data_store"`
}

// given a path to a config file parse it from json
func (c cabbyConfig) parse(file string) (pc cabbyConfig) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		fail.Panic(err)
	}

	if err = json.Unmarshal(b, &pc); err != nil {
		fail.Panic(err)
	}

	return
}
