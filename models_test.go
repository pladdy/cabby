package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
)

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

/* Discovery */

func TestParseDiscoveryResource(t *testing.T) {
	result := string(parseDiscoveryResource(DiscoveryResourceFile))

	if len(result) == 0 {
		t.Error("Got:", result, "Expected: length > 0")
	}
}

func TestParseDiscoveryResourceNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	_ = string(parseDiscoveryResource("foo/bar"))
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
