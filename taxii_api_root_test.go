package main

import (
	"encoding/json"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestHandleTaxiiAPIRoot(t *testing.T) {
	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])
	status, result := handlerTest(handleTaxiiAPIRoot, "GET", noPortHost, nil)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleTaxiiAPIRootNoconfig(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleTaxiiAPIRootNotDefined(t *testing.T) {
	config = cabbyConfig{}.parse("test/config/no_discovery_config.json")
	defer reloadTestConfig()

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
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
		// create a config struct with an API Root and corresponding API Root Map
		a := taxiiAPIRoot{
			Title:            "test",
			Description:      "test api root",
			Versions:         []string{"taxii-2.0"},
			MaxContentLength: 1}

		c := cabbyConfig{APIRootMap: map[string]taxiiAPIRoot{test.rootMapEntry: a}}

		c.Discovery = taxiiDiscovery{APIRoots: []string{test.apiRoot}}

		result := c.validAPIRoot(test.apiRoot)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
