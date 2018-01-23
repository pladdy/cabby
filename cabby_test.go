package main

import (
	"crypto/tls"
	"encoding/json"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"testing"
	"time"
)

const (
	discoveryURL = "https://localhost:1234/taxii"
	apiRootURL   = "https://localhost:1234/api_root"
	testUser     = "test@cabby.com"
	testPass     = "test"
)

/* helpers */

func attemptRequest(c *http.Client, r *http.Request) (*http.Response, error) {
	MaxTries := 3

	for i := 0; i < MaxTries; i++ {
		res, err := c.Do(r)
		if err != err {
			log.Fatal(err)
		}
		if res != nil {
			return res, err
		}
		logWarn.Println("Web server for test not responding, waiting...")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return nil, errors.New("Failed to get request")
}

func get(u string) string {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/testing_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/testing_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	c := cabbyConfig{}.parse(configPath)
	server := newCabby(c)
	go func() {
		server.ListenAndServeTLS(c.SSLCert, c.SSLKey)
	}()

	// set up client with TLS configured
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	client := &http.Client{Transport: tr}

	req, err := http.NewRequest("GET", u, nil)
	if err != nil {
		log.Fatal(err)
	}
	req.SetBasicAuth(testUser, testPass)

	res, err := attemptRequest(client, req)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != err {
		log.Fatal(err)
	}

	server.Close()
	return string(body)
}

/* tests */

func TestMain(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/different_port_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/different_port_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	go func() {
		main()
	}()

	// rename files back in reverse (order matters or you clobber the files)
	time.Sleep(100 * time.Millisecond)
}

func TestMainDiscovery(t *testing.T) {
	response := get(discoveryURL)
	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.Discovery)

	if response != string(expected) {
		t.Error("Got:", response, "Expected:", string(expected))
	}
}

func TestMainAPIRoot(t *testing.T) {
	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	response := get(apiRootURL)
	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])

	if response != string(expected) {
		t.Error("Got:", response, "Expected:", string(expected))
	}
}

func TestRegisterAPIRootInvalidURL(t *testing.T) {
	recovered := false

	defer func() {
		if err := recover(); err == nil {
			t.Error("Failed to recover")
		}
		recovered = true
	}()

	mockHandler := http.NewServeMux()
	registerAPIRoot("", mockHandler)

	if recovered != true {
		t.Error("Expected: 'recovered' to be true")
	}
}
