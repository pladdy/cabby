package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
)

/* helpers */

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		log.Fatal("Failed to rename file:", from, "to:", to)
	}
}

// define a handler function type for handler testing
type handlerFn func(http.ResponseWriter, *http.Request)

// handlerTest is a function to handle generic testing of handlers
// it takes a handler function to call with a url; it returns the stuats code
// and response as a string
func handlerTest(h handlerFn, u string) (int, string) {
	req := httptest.NewRequest("GET", u, nil)
	res := httptest.NewRecorder()
	h(res, req)

	b, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(b)
}

/* auth tests */

func TestBasicAuth(t *testing.T) {
	tests := []struct {
		user     string
		pass     string
		expected int
	}{
		{"pladdy", "pants", 200},
		{"simon", "says", 401},
	}

	testHandlerAuth := basicAuth(
		func(w http.ResponseWriter, r *http.Request) {
			io.WriteString(w, "test")
		})

	for _, test := range tests {
		req := httptest.NewRequest("GET", discoveryURL, nil)
		req.SetBasicAuth(test.user, test.pass)
		res := httptest.NewRecorder()
		testHandlerAuth(res, req)

		if res.Code != test.expected {
			t.Error("Got:", res.Code, "Expected:", test.expected)
		}
	}
}

func TestValidate(t *testing.T) {
	tests := []struct {
		user     string
		pass     string
		expected bool
	}{
		{testUser, testPass, true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		actual := validated(test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
		}
	}
}

/* handleDiscovery */

func TestHandleDiscovery(t *testing.T) {
	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.Discovery)
	status, result := handlerTest(handleDiscovery, discoveryURL)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleDiscoveryNoconfig(t *testing.T) {
	renameFile(configPath, configPath+".testing")

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	renameFile(configPath+".testing", configPath)
}

func TestHandleDiscoveryNotDefined(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/no_discovery_config.json", configPath)

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	// rename files back in reverse (order matters or you clobber the files)
	renameFile(configPath, "test/no_discovery_config.json")
	renameFile(configPath+".testing", configPath)
}

/* handleAPIRoot */

func TestHandleAPIRoot(t *testing.T) {
	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])
	status, result := handlerTest(handleAPIRoot, noPortHost)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleAPIRootNoconfig(t *testing.T) {
	renameFile(configPath, configPath+".testing")

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	renameFile(configPath+".testing", configPath)
}

func TestHandleAPIRootNotDefined(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/no_discovery_config.json", configPath)

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	// rename files back in reverse (order matters or you clobber the files)
	renameFile(configPath, "test/no_discovery_config.json")
	renameFile(configPath+".testing", configPath)
}

/* undefined request */

func TestHandleUndefinedRequest(t *testing.T) {
	status, result := handlerTest(handleUndefinedRequest, "/nobody-home")
	if status != 404 {
		t.Error("Got:", status, "Expected: 404", "Response:", result)
	}
}

/* handler helper tests */

func TestUrlWithNoPort(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"https://localhost:1234/api_root", "https://localhost/api_root"},
		{"https://localhost/api_root", "https://localhost/api_root"},
		{"/api_root", "https://localhost/api_root"},
	}

	for _, test := range tests {
		u, _ := url.Parse(test.host)
		result := urlWithNoPort(u)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(handleDiscovery))
	defer ts.Close()

	res, err := http.Get(ts.URL)
	if err != nil {
		log.Fatal(err)
	}

	result := res.Header["Content-Type"][0]
	expected := taxiiContentType

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestResourceToJSON(t *testing.T) {
	tests := []struct {
		resource interface{}
		expected string
	}{
		{taxiiAPIRoot{Title: "apiRoot", Description: "apiRoot", Versions: []string{"test-1.0"}, MaxContentLength: 1},
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
