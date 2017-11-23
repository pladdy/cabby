package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const DiscoveryURL = "http://localhost:1234/taxii"

var testResource = DiscoveryResource{
	"Test Discovery",
	"This is a test discovery resource",
	"pladdy",
	"https://test.com/api1",
	[]string{"https://test.com/api2", "https://test.com/api3"}}

func getResponseBody(url string) string {
	res, err := http.Get(url)
	if err != err {
		log.Fatal(err)
	}
	defer res.Body.Close()

	body, err := ioutil.ReadAll(res.Body)
	if err != err {
		log.Fatal(err)
	}
	return string(body)
}

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Fatal(err)
	}
	json := string(b)
	return json
}

/* tests */

func TestHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	result := getResponseBody(ts.URL)
	expected := resourceToJSON(testResource)

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	result := res.Header["Content-Type"][0]
	expected := TAXIIContentType

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

// Run the webserver and test it
func TestMain(t *testing.T) {
	go func() {
		main()
	}()

	result := getResponseBody(DiscoveryURL)
	expected := resourceToJSON(testResource)

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}
