package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
)

func TestParseConfig(t *testing.T) {
	cs := configs{}.parse("config/cabby.example.json")

	if cs["development"].Host != "localhost" {
		t.Error("Got:", "localhost", "Expected:", "localhost")
	}
	if cs["development"].Port != 1234 {
		t.Error("Got:", strconv.Itoa(1234), "Expected:", strconv.Itoa(1234))
	}
}

func TestParseConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	_ = configs{}.parse("foo/bar")
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

	ioutil.WriteFile(invalidJSON, []byte("invalid"), 0644)
	configs{}.parse(invalidJSON)
	t.Error("Failed to panic with an unknown resource")
}
