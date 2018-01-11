package main

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

/* helpers */

// define a handler function type for handler testing
type handlerFn func(http.ResponseWriter, *http.Request)

// handle generic testing of handlers.  It takes a handler function to call with a url;
// it returns the status code and response as a string
func handlerTest(h handlerFn, m, u string) (int, string) {
	req := httptest.NewRequest(m, u, nil)
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

/* handleTaxiiAPIRoot */

func TestHandleTaxiiAPIRoot(t *testing.T) {
	u, _ := url.Parse(apiRootURL)
	noPortHost := urlWithNoPort(u)

	config := cabbyConfig{}.parse(configPath)
	expected, _ := json.Marshal(config.APIRootMap[noPortHost])
	status, result := handlerTest(handleTaxiiAPIRoot, "GET", noPortHost)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	if result != string(expected) {
		t.Error("Got:", result, "Expected:", string(expected))
	}
}

func TestHandleTaxiiAPIRootNoconfig(t *testing.T) {
	renameFile(configPath, configPath+".testing")

	defer func() {
		renameFile(configPath+".testing", configPath)
	}()

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleTaxiiAPIRootNotDefined(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/no_discovery_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/no_discovery_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	req := httptest.NewRequest("GET", apiRootURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiAPIRoot(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

/* handleTaxiiCollection */

func TestHandleTaxiiCollectionCreate(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/testing_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/testing_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	// set up URL
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Error(err)
	}

	q := u.Query()
	q.Set("title", t.Name())
	q.Set("description", "a description")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollection, "POST", u.String())

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	// check on record
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB
	s, err := newSQLiteDB(c)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	var title string
	err = s.db.QueryRow("select title from taxii_collection where title = '" + t.Name() + "'").Scan(&title)
	if err != nil {
		t.Error(err)
	}

	if title != t.Name() {
		t.Error("Got:", title, "Expected:", t.Name())
	}
}

func TestHandleTaxiiCollectionCreateBadID(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/testing_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/testing_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	// set up URL
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Error(err)
	}

	q := u.Query()
	q.Set("id", "fail")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollection, "POST", u.String())

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}

	// verify no record exists
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB
	s, err := newSQLiteDB(c)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	var title string
	err = s.db.QueryRow("select id from taxii_collection where id = 'fail'").Scan(&title)
	if err == nil {
		t.Fatal("Should be no record created")
	}
}

func TestHandleTaxiiCollectionCreateBadParse(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/testing_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/testing_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	req := httptest.NewRequest("POST", "https://localhost/api_root/collections", nil)

	// change body to nil to trigger a parse error once the handler tries to parse the form elements in the url
	req.Body = nil

	res := httptest.NewRecorder()
	handleTaxiiCollection(res, req)

	if res.Code != 400 {
		t.Error("Got:", res.Code, "Expected:", 400)
	}

	req = httptest.NewRequest("CUSTOM", "https://localhost/api_root/collections", nil)
	res = httptest.NewRecorder()
	handleTaxiiCollection(res, req)

	if res.Code != 400 {
		t.Error("Got:", res.Code, "Expected:", 400)
	}
}

func TestHandleTaxiiCollectionCreateInvalidDB(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/no_datastore_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/no_datastore_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	// set up URL
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Error(err)
	}

	q := u.Query()
	q.Set("title", "a title")
	q.Set("description", "a description")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollection, "POST", u.String())

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
}

/* handleTaxiiDiscovery */

func TestHandleDiscovery(t *testing.T) {
	config := cabbyConfig{}.parse(configPath)
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
	renameFile(configPath, configPath+".testing")

	defer func() {
		renameFile(configPath+".testing", configPath)
	}()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleDiscoveryNotDefined(t *testing.T) {
	renameFile(configPath, configPath+".testing")
	renameFile("test/config/no_discovery_config.json", configPath)

	defer func() {
		renameFile(configPath, "test/config/no_discovery_config.json")
		renameFile(configPath+".testing", configPath)
	}()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

/* undefined request */

func TestHandleUndefinedRequest(t *testing.T) {
	status, result := handlerTest(handleUndefinedRequest, "GET", "/nobody-home")
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
	ts := httptest.NewServer(http.HandlerFunc(handleTaxiiDiscovery))
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
