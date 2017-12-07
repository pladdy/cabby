package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
)

/* API Roots */

func TestAPIRootStruct(t *testing.T) {
	config := Config{}.parse("config/cabby.example.json")
	testRoot := "https://localhost/api_root"

	if len(config.APIRootMap[testRoot].Title) == 0 {
		t.Error("field not set in API root")
	}
	if len(config.APIRootMap[testRoot].Description) == 0 {
		t.Error("field not set in API root")
	}
	if len(config.APIRootMap[testRoot].Versions) == 0 {
		t.Error("field not set in API root")
	}
	if config.APIRootMap[testRoot].MaxContentLength <= 0 {
		t.Error("field not set in API root")
	}
}

func TestAPIRootVerify(t *testing.T) {
	tests := []struct {
		apiRoot      string
		rootMapEntry string
		expected     bool
	}{
		{"https://localhost/api_test", "https://localhost/api_test", true},
		{"https://localhost/api_fail", "https://localhost/api_test", false},
	}

	for _, test := range tests {
		// create a Config struct with an API Root and corresponding API Root Map
		a := APIRoot{
			Title:            "test",
			Description:      "test api root",
			Versions:         []string{"1.0"},
			MaxContentLength: 1}

		c := Config{APIRootMap: map[string]APIRoot{test.rootMapEntry: a}}

		c.Discovery = DiscoveryResource{APIRoots: []string{test.apiRoot}}

		result := c.validAPIRoot(test.apiRoot)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

/* Config */

func TestParseConfig(t *testing.T) {
	config := Config{}.parse("config/cabby.example.json")

	if config.Host != "localhost" {
		t.Error("Got:", "localhost", "Expected:", "localhost")
	}
	if config.Port != 1234 {
		t.Error("Got:", strconv.Itoa(1234), "Expected:", strconv.Itoa(1234))
	}
}

func TestParseConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	_ = Config{}.parse("foo/bar")
	t.Error("Failed to panic with an unknown resource")
}

func TestParseConfigInvalidJSON(t *testing.T) {
	invalidJSON := "invalid.json"

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
			os.Remove(invalidJSON)
		}
	}()

	_ = ioutil.WriteFile(invalidJSON, []byte("invalid"), 0644)
	_ = Config{}.parse(invalidJSON)
	t.Error("Failed to panic with an unknown resource")
}

/* Error */

func TestErrorMessage(t *testing.T) {
	testError := Error{Title: "Test title", HTTPStatus: 404}
	result := testError.Message()
	expected := `{"title":"Test title","http_status":"404"}`

	if testError.Message() != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}
