package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHandleDiscovery(t *testing.T) {
	expected, _ := json.Marshal(config.Discovery)
	status, result := handlerTest(handleTaxiiDiscovery, "GET", discoveryURL)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleDiscoveryNoconfig(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleDiscoveryNotDefined(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}
