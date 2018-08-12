package http

import (
	"testing"

	cabby "github.com/pladdy/cabby2"
)

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
