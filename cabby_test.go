package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

const (
	DiscoveryURL = "http://localhost:1234/taxii"
	TestUser     = "pladdy"
	TestPass     = "pants"
)

/* helpers */

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

func TestDiscoveryHandler(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleDiscovery))
	defer ts.Close()

	result := getResponseBody(ts.URL)
	expected := resourceToJSON(testResource)

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleDiscovery))
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

	client := &http.Client{}

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

	expected := resourceToJSON(testResource)

	if string(body) != expected {
		t.Error("Got:", string(body), "Expected:", expected)
	}
}

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		user     string
		pass     string
		expected int
	}{
		{"pladdy", "pants", 200},
		{"simon", "says", 401},
	}

	testHandlerAuth := basicAuth(
		func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "test")
		})

	for _, test := range tests {
		req := httptest.NewRequest("GET", DiscoveryURL, nil)
		req.SetBasicAuth(test.user, test.pass)
		res := httptest.NewRecorder()
		testHandlerAuth(res, req)

		if res.Code != test.expected {
			t.Error("Got:", res.Code, "Expected:", test.expected)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		user     string
		pass     string
		expected bool
	}{
		{"pladdy", "pants", true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		actual := validate(test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
		}
	}
}
