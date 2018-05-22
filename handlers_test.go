package main

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetToken(t *testing.T) {
	tests := []struct {
		url      string
		index    int
		expected string
	}{
		{"/api_root/collections/collection_id/objects/stix_id", 0, ""},
		{"/api_root/collections/collection_id/objects/stix_id", 1, "api_root"},
		{"/api_root/collections/collection_id/objects/stix_id", 3, "collection_id"},
		{"/api_root/collections/collection_id/objects/stix_id", 5, "stix_id"},
		{"/api_root/collections/collection_id/objects/stix_id", 7, ""},
	}

	for _, test := range tests {
		result := getToken(test.url, test.index)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestGetAPIRoot(t *testing.T) {
	tests := []struct {
		urlPath  string
		expected string
	}{
		{"/api_root/collection/1234", "api_root"},
		{"/api_root/collections", "api_root"},
	}

	for _, test := range tests {
		result := getAPIRoot(test.urlPath)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
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

func TestSetTaxiiFilter(t *testing.T) {
	tests := []struct {
		input    string
		expected string
		err      error
	}{
		{"2016-04-06T20:03:48.000Z", "2016-04-06T20:03:48Z", nil},
		{"2016-04-06T20:03:48.0122Z", "2016-04-06T20:03:48.0122Z", nil},
		{"2016-04-06 20:03:48.0122Z", "2016-04-06 20:03:48.0122Z", errors.New("not nil")},
	}

	for _, test := range tests {
		tf := taxiiFilter{}
		err := tf.setAddedAfter(test.input)

		if err == nil && tf.addedAfter != test.expected {
			t.Error("Got:", tf.addedAfter, "Expected:", test.expected)
		}

		if test.err != nil && err == nil {
			t.Error("Got:", err, "Expected:", test.err)
		}
	}
}

func TestNewTaxiiRange(t *testing.T) {
	invalidRange := taxiiRange{first: -1, last: -1}

	tests := []struct {
		input       string
		resultRange taxiiRange
		isError     bool
	}{
		{"items 0-10", taxiiRange{first: 0, last: 10}, false},
		{"items 0 10", invalidRange, true},
		{"items 10", invalidRange, true},
		{"", invalidRange, false},
	}

	for _, test := range tests {
		result, err := newTaxiiRange(test.input)
		if result != test.resultRange {
			t.Error("Got:", result, "Expected:", test.resultRange)
		}

		if err != nil && test.isError == false {
			t.Error("Got:", err, "Expected: no error")
		}
	}
}

func TestTaxiiRangeValid(t *testing.T) {
	tests := []struct {
		tr       taxiiRange
		expected bool
	}{
		{taxiiRange{first: 1, last: 0}, false},
		{taxiiRange{first: 0, last: 0}, true},
	}

	for _, test := range tests {
		result := test.tr.Valid()
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestTaxiiRangeString(t *testing.T) {
	tests := []struct {
		tr       taxiiRange
		expected string
	}{
		{taxiiRange{first: 0, last: 0}, "items 0-0"},
		{taxiiRange{first: 0, last: 0, total: 50}, "items 0-0/50"},
	}

	for _, test := range tests {
		result := test.tr.String()
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

func TestTakeRequestRange(t *testing.T) {
	req := httptest.NewRequest("GET", "/test", nil)

	result := takeRequestRange(req)
	expected := taxiiRange{}

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
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
