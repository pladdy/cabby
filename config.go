package main

import (
	"encoding/json"
	"io/ioutil"
)

var config cabbyConfig

type cabbyConfig struct {
	Host       string
	Port       int
	SSLCert    string                  `json:"ssl_cert"`
	SSLKey     string                  `json:"ssl_key"`
	DataStore  map[string]string       `json:"data_store"`
	Discovery  taxiiDiscovery          `json:"discovery"`
	APIRootMap map[string]taxiiAPIRoot `json:"api_root_map"`
}

func (c *cabbyConfig) discoveryDefined() bool {
	if c.Discovery.Title == "" {
		return false
	}
	return true
}

// check if given api root is in the API Root Map of a config
func (c cabbyConfig) inAPIRootMap(ar string) (r bool) {
	for k := range c.APIRootMap {
		if ar == k {
			r = true
		}
	}
	return
}

// check if given api root is in the API Roots of a config
func (c cabbyConfig) inAPIRoots(ar string) (r bool) {
	for _, v := range c.Discovery.APIRoots {
		if ar == v {
			r = true
		}
	}
	return
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

// validate that an api root has a same definition in apiRootMap
func (c cabbyConfig) validAPIRoot(a string) bool {
	return c.inAPIRoots(a) && c.inAPIRootMap(a)
}
