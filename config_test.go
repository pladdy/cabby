package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestParseConfig(t *testing.T) {
	config = cabbyConfig{}.parse("config/cabby.example.json")
	defer loadTestConfig()

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

	_ = cabbyConfig{}.parse("foo/bar")
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
	_ = cabbyConfig{}.parse(invalidJSON)
	t.Error("Failed to panic with an unknown resource")
}
