package http

import (
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

// func TestResourceToJSON(t *testing.T) {
// 	tests := []struct {
// 		resource interface{}
// 		expected string
// 	}{
// 		{taxiiAPIRoot{Title: "apiRoot", Description: "apiRoot", Versions: []string{"test-1.0"}, MaxContentLength: 1},
// 			`{"title":"apiRoot","description":"apiRoot","versions":["test-1.0"],"max_content_length":1}`},
// 	}
//
// 	for _, test := range tests {
// 		result := resourceToJSON(test.resource)
//
// 		if result != test.expected {
// 			t.Error("Got:", result, "Expected:", test.expected)
// 		}
// 	}
// }

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
