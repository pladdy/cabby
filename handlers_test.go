package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHeaders(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	h := handleTaxiiDiscovery(ts, testConfig().Port)
	s := httptest.NewServer(http.HandlerFunc(h))
	defer s.Close()

	res, err := http.Get(s.URL)
	if err != nil {
		log.Fatal(err)
	}

	result := res.Header["Content-Type"][0]
	expected := taxiiContentType

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestRequireAcceptStix(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
	}
	mockHandler = withAcceptStix(mockHandler)

	tests := []struct {
		acceptHeader string
		responseCode int
	}{
		{"application/vnd.oasis.stix+json; version=2.0", http.StatusOK},
		{"application/vnd.oasis.stix+json", http.StatusOK},
		{"application/vnd.oasis.stix+json;verion=2.0", http.StatusOK},
		{"", http.StatusUnsupportedMediaType},
		{"application/vnd.oasis.stix+jsonp", http.StatusUnsupportedMediaType},
		{"application/vnd.oasis.stix+jsonp; version=3.0", http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		mockHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", http.StatusOK)
		}
	}
}

func TestRequireAcceptTaxii(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
	}
	mockHandler = withAcceptTaxii(mockHandler)

	tests := []struct {
		acceptHeader string
		responseCode int
	}{
		{"application/vnd.oasis.taxii+json; version=2.0", http.StatusOK},
		{"application/vnd.oasis.taxii+json", http.StatusOK},
		{"application/vnd.oasis.taxii+json;verion=2.0", http.StatusOK},
		{"", http.StatusUnsupportedMediaType},
		{"application/vnd.oasis.taxii+jsonp", http.StatusUnsupportedMediaType},
		{"application/vnd.oasis.taxii+jsonp; version=3.0", http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		mockHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", http.StatusOK)
		}
	}
}
