package main

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

/* helpers */

// handle generic testing of handlers.  It takes a handler function to call with a url;
// it returns the status code and response as a string
func handlerTest(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	var req *http.Request

	if b != nil {
		req = httptest.NewRequest("POST", url, b)
	} else {
		req = httptest.NewRequest(method, url, nil)
	}

	ctx := context.WithValue(context.Background(), userName, testUser)
	req = req.WithContext(ctx)
	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

func TestRequireAcceptTaxii(t *testing.T) {
	mockHandler := func(w http.ResponseWriter, r *http.Request) {
		accept := r.Header.Get("Accept")
		io.WriteString(w, fmt.Sprintln("Accept Header: %v", accept))
	}
	mockHandler = withAcceptTaxii(mockHandler)

	tests := []struct {
		acceptHeader string
		responseCode int
	}{
		{"application/vnd.oasis.taxii+json; version=2.0", 200},
		{"application/vnd.oasis.taxii+json", 200},
		{"application/vnd.oasis.taxii+json;verion=2.0", 200},
		{"", 415},
		{"application/vnd.oasis.taxii+jsonp", 415},
		{"application/vnd.oasis.taxii+jsonp; version=3.0", 415},
	}

	for _, test := range tests {
		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		mockHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected: 200")
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

/* undefined request */

func TestHandleUndefinedRequest(t *testing.T) {
	status, result := handlerTest(handleUndefinedRequest, "GET", "/nobody-home", nil)
	if status != 404 {
		t.Error("Got:", status, "Expected: 404", "Response:", result)
	}
}

func TestHeaders(t *testing.T) {
	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		t.Fatal(err)
	}
	defer ts.disconnect()

	h := handleTaxiiDiscovery(ts)
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

func TestUrlWithNoPort(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"https://localhost:1234/api_root", "https://localhost/api_root"},
		{"https://localhost/api_root", "https://localhost/api_root"},
		{"/api_root", "https://localhost/api_root"},
	}

	for _, test := range tests {
		u, _ := url.Parse(test.host)
		result := urlWithNoPort(u)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
