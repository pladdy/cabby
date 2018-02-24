package main

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
)

var config Config

type Config struct {
	Host      string
	Port      int
	SSLCert   string            `json:"ssl_cert"`
	SSLKey    string            `json:"ssl_key"`
	DataStore map[string]string `json:"data_store"`
}

// given a path to a config file parse it from json
func (c Config) parse(file string) (pc Config) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.WithFields(log.Fields{
			"file":  file,
			"error": err,
		}).Panic("Can't parse config file")
	}

	if err = json.Unmarshal(b, &pc); err != nil {
		log.WithFields(log.Fields{
			"file":  file,
			"error": err,
		}).Panic("Can't unmarshal JSON")
	}

	return
}
