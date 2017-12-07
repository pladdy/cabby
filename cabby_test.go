package main

import (
	"crypto/tls"
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"testing"
)

const (
	DiscoveryURL = "https://localhost:1234/taxii"
	APIRootURL   = "https://localhost:1234/api_root"
	TestUser     = "pladdy"
	TestPass     = "pants"
)

// Run the webserver and test it
func TestMain(t *testing.T) {
	go func() {
		main()
	}()

	// set up client with TLS configured
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", DiscoveryURL, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(TestUser, TestPass)

	res, err := client.Do(req)
	if err != err {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != err {
		log.Fatal(err)
	}

	config := Config{}.parse(ConfigPath)
	expected, _ := json.Marshal(config.Discovery)

	if string(body) != string(expected) {
		t.Error("Got:", string(body), "Expected:", string(expected))
	}
}
