package http

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

func TestNoResources(t *testing.T) {
	resources := 0

	result := noResources(resources)
	if result != true {
		t.Error("Expected true")
	}
}

func TestObjectsToEnvelopeMore(t *testing.T) {
	e := objectsToEnvelope([]stones.Object{}, cabby.Page{Total: 1})
	expected := true
	if e.More != expected {
		t.Error("Got:", e.More, "Expected:", expected)
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
