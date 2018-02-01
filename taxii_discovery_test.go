package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHandleDiscovery(t *testing.T) {
	status, result := handlerTest(handleTaxiiDiscovery, "GET", discoveryURL, nil)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	var resultTC taxiiDiscovery
	err := json.Unmarshal([]byte(result), &resultTC)
	if err != nil {
		t.Fatal(err)
	}

	expected := "https://localhost:1234/taxii/"
	if resultTC.Default != expected {
		t.Error("Got:", resultTC.Default, "Expected:", expected)
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
