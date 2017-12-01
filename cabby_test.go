package main

import (
	"crypto/tls"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
)

const (
	DiscoveryURL = "https://localhost:1234/taxii"
	TestUser     = "pladdy"
	TestPass     = "pants"
)

func TestDiscoveryHandler(t *testing.T) {
	req := httptest.NewRequest("GET", DiscoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 200 {
		t.Error("Got:", res.Code, "Expected:", 200)
	}

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	result := string(b)
	expected := string(parseDiscoveryResource(DiscoveryResourceFile))

	if result != expected {
		t.Error("Got:", string(result), "Expected:", string(expected))
	}
}

func TestDiscoveryHandlerNoResource(t *testing.T) {
	oldPath := DiscoveryResourceFile
	newPath := oldPath + ".testing"

	// rename the resource so it can't be found
	err := os.Rename(oldPath, newPath)
	if err != nil {
		log.Fatal("Failed to rename test file:", oldPath)
	}

	req := httptest.NewRequest("GET", DiscoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
	err = os.Rename(newPath, oldPath)
	if err != nil {
		log.Fatal("Failed to rename test file:", newPath)
	}
}

func TestStrictTransportSecurity(t *testing.T) {
	resultKey, resultValue := strictTransportSecurity()
	expectedKey, expectedValue := "Strict-Transport-Security", "max-age=63072000; includeSubDomains"

	if resultKey != expectedKey {
		t.Error("Got:", resultKey, "Expected:", expectedKey)
	}

	if resultValue != expectedValue {
		t.Error("Got:", resultValue, "Expected:", expectedValue)
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

	expected := string(parseDiscoveryResource(DiscoveryResourceFile))

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
		{TestUser, TestPass, true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		actual := validate(test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
		}
	}
}
