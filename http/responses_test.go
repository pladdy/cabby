package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pladdy/cabby"
)

func TestNoResources(t *testing.T) {
	res := httptest.NewRecorder()
	resources := 0
	cr := cabby.Range{Set: true}

	result := noResources(res, resources, cr)
	if result != true {
		t.Error("Expected true")
	}
}

func TestResourceToJSON(t *testing.T) {
	tests := []struct {
		resource cabby.APIRoot
		expected string
	}{
		{cabby.APIRoot{Title: "apiRoot", Description: "apiRoot", Versions: []string{"test-1.0"}, MaxContentLength: 1},
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

func TestWrite(t *testing.T) {
	content := "foo"

	tests := []struct {
		r        *http.Request
		expected string
	}{
		{httptest.NewRequest("GET", testDiscoveryURL, nil), content},
		{httptest.NewRequest("HEAD", testDiscoveryURL, nil), ""},
	}

	for _, test := range tests {
		w := httptest.NewRecorder()
		write(w, test.r, content)
		result, _ := ioutil.ReadAll(w.Body)

		if string(result) != test.expected {
			t.Error("Got:", string(result), "Expected:", test.expected)
		}
	}
}
