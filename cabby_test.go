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
	testUser     = "pladdy"
	testPass     = "pants"
)

func init() {
	go func() {
		main()
	}()
}

/* helpers */

func attempt(c *http.Client, r *http.Request) (*http.Response, error) {
	for i := 0; i < 3; i++ {
		res, err := c.Do(r)
		if err != err {
			log.Fatal(err)
		}
		if res != nil {
			return res, err
		}
		warn.Println("Web server for test not responding, waiting...")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return nil, errors.New("Failed to get request")
}

func get(u string) string {
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

	res, err := attempt(client, req)
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != err {
		log.Fatal(err)
	}

	return string(body)
}

/* test */
func TestMainDiscovery(t *testing.T) {
	response := get(discoveryURL)
	config := config{}.parse(configPath)
	expected, _ := json.Marshal(config.Discovery)

	if response != string(expected) {
		t.Error("Got:", response, "Expected:", string(expected))
	}
}

func TestMainAPIRoot(t *testing.T) {
	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	response := get(apiRootURL)
	config := config{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])

	if response != string(expected) {
		t.Error("Got:", response, "Expected:", string(expected))
	}
}
