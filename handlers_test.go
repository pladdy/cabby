package main

import (
	"context"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

/* helpers */

// define a handler function type for handler testing
type handlerFn func(http.ResponseWriter, *http.Request)

// handle generic testing of handlers.  It takes a handler function to call with a url;
// it returns the status code and response as a string
func handlerTest(h handlerFn, m, u string) (int, string) {
	req := httptest.NewRequest(m, u, nil)
	ctx := context.WithValue(context.Background(), userName, testUser)
	req = req.WithContext(ctx)

	res := httptest.NewRecorder()
	h(res, req)

	b, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(b)
}

/* auth tests */

func TestValidateUser(t *testing.T) {
	tests := []struct {
		user     string
		pass     string
		expected bool
	}{
		{testUser, testPass, true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		_, actual := validateUser(test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
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
	status, result := handlerTest(handleUndefinedRequest, "GET", "/nobody-home")
	if status != 404 {
		t.Error("Got:", status, "Expected: 404", "Response:", result)
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleTaxiiDiscovery))
	defer ts.Close()

	res, err := http.Get(ts.URL)
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
