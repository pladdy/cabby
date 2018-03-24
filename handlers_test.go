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

func TestAPIRoot(t *testing.T) {
	tests := []struct {
		urlPath  string
		expected string
	}{
		{"/api_root/collection/1234", "api_root"},
		{"/api_root/collections", "api_root"},
	}

	for _, test := range tests {
		result := apiRoot(test.urlPath)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestWithLoggingUnauthorized(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	// omit adding a user to the context
	res := httptest.NewRecorder()
	withRequestLogging(handleTaxiiDiscovery(ts, config.Port))(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Error("Got:", res.Code, "Expected: unauthorized")
	}
}

func TestHandleUndefinedRequest(t *testing.T) {
	status, result := handlerTest(handleUndefinedRequest, "GET", "/nobody-home", nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound, "Response:", result)
	}
}

func TestHeaders(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	h := handleTaxiiDiscovery(ts, config.Port)
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

/* handler helper tests */

func TestLastURLPathToken(t *testing.T) {
	tests := []struct {
		path     string
		expected string
	}{
		{"/collections/", "collections"},
		{"/collections/someId", "someId"},
	}

	for _, test := range tests {
		result := lastURLPathToken(test.path)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestRecoverFromPanic(t *testing.T) {
	w := httptest.NewRecorder()
	defer recoverFromPanic(w)
	panic("test")
}

func TestResourceToJSON(t *testing.T) {
	tests := []struct {
		resource interface{}
		expected string
	}{
		{taxiiAPIRoot{Title: "apiRoot", Description: "apiRoot", Versions: []string{"test-1.0"}, MaxContentLength: 1},
			`{"title":"apiRoot","description":"apiRoot","versions":["test-1.0"],"max_content_length":1}`},
	}

	for _, test := range tests {
		result := resourceToJSON(test.resource)

		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestResourceToJSONFail(t *testing.T) {
	recovered := false

	defer func() {
		if err := recover(); err == nil {
			t.Error("Failed to recover:", err)
		}
		recovered = true
	}()

	c := make(chan int)
	result := resourceToJSON(c)

	if recovered != true {
		t.Error("Got:", result, "Expected: 'recovered' to be true")
	}
}
