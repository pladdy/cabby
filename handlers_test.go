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
// generic handler testing is make sure handler returns a 200 and
// an expected output
func handlerTest(h handlerFn, u string) (status int, expected string) {
	req := httptest.NewRequest("GET", u, nil)
	res := httptest.NewRecorder()
	h(res, req)

	b, _ := ioutil.ReadAll(res.Body)
	return 200, string(b)
}

/* discovery handler tests */

func TestHandleDiscovery(t *testing.T) {
	config := Config{}.parse(ConfigPath)
	expected, _ := json.Marshal(config.Discovery)
	status, result := handlerTest(handleDiscovery, DiscoveryURL)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleDiscoveryNoConfig(t *testing.T) {
	renameFile(ConfigPath, ConfigPath+".testing")

	req := httptest.NewRequest("GET", DiscoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	renameFile(ConfigPath+".testing", ConfigPath)
}

func TestHandleDiscoveryNoResourceDefined(t *testing.T) {
	renameFile(ConfigPath, ConfigPath+".testing")
	renameFile("test/no_discovery_config.json", ConfigPath)

	req := httptest.NewRequest("GET", DiscoveryURL, nil)
	res := httptest.NewRecorder()
	handleDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	// rename files back in reverse (order matters or you clobber the files)
	renameFile(ConfigPath, "test/no_discovery_config.json")
	renameFile(ConfigPath+".testing", ConfigPath)
}

/* handleAPIRoot tests */

func TestHandleAPIRoot(t *testing.T) {
	u, _ := url.Parse(APIRootURL)
	noPort := removePort(u)

	config := Config{}.parse(ConfigPath)
	expected, _ := json.Marshal(config.APIRootMap[noPort])
	status, result := handlerTest(handleAPIRoot, noPort)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleAPIRootNoConfig(t *testing.T) {
	renameFile(ConfigPath, ConfigPath+".testing")

	req := httptest.NewRequest("GET", APIRootURL, nil)
	res := httptest.NewRecorder()
	handleAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	renameFile(ConfigPath+".testing", ConfigPath)
}

func TestHandleAPIRootNoResourceDefined(t *testing.T) {
	renameFile(ConfigPath, ConfigPath+".testing")
	renameFile("test/no_discovery_config.json", ConfigPath)

	req := httptest.NewRequest("GET", APIRootURL, nil)
	res := httptest.NewRecorder()
	handleAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}

	// rename files back in reverse (order matters or you clobber the files)
	renameFile(ConfigPath, "test/no_discovery_config.json")
	renameFile(ConfigPath+".testing", ConfigPath)
}

func TestRemovePort(t *testing.T) {
	tests := []struct {
		host     string
		expected string
	}{
		{"https://localhost:1234/api_root", "https://localhost/api_root"},
		{"https://localhost/api_root", "https://localhost/api_root"},
	}

	for _, test := range tests {
		u, _ := url.Parse(test.host)
		result := removePort(u)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

/* misc tests */

func TestStrictTransportSecurity(t *testing.T) {
	resultKey, resultValue := strictTransportSecurity()
	expectedKey, expectedValue := "Strict-Transport-Security", "max-age=63072000; includeSubDomains"

	if resultKey != expectedKey {
		t.Error("Got:", resultKey, "Expected:", expectedKey)
	}

	if resultValue != expectedValue {
		t.Error("Got:", resultValue, "Expected:", expectedValue)
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
	expected := TAXIIContentType

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

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
		req := httptest.NewRequest("GET", DiscoveryURL, nil)
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
		{TestUser, TestPass, true},
		{"simon", "says", false},
	}

	for _, test := range tests {
		actual := validate(test.user, test.pass)
		if actual != test.expected {
			t.Error("Got:", actual, "Expected:", test.expected)
		}
	}
}
