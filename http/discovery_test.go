package http

import (
	"encoding/json"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestDiscoveryHandlerServeHTTP(t *testing.T) {
	// TODO: create a mock DiscoveryService
	s := DiscoveryService{DB: ds.DB}
	h := DiscoveryHandler{DiscoveryService: s}

	status, result := handlerTest(h(testPort), "GET", testDiscoveryURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.Discovery
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}

	if discovery.Title != testDiscovery.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != testDiscovery.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
	if discovery.Contact != testDiscovery.Contact {
		t.Error("Got:", discovery.Contact, "Expected:", expected.Contact)
	}
	if discovery.Default != expected.Default {
		t.Error("Got:", discovery.Default, "Expected:", expected.Default)
	}
}

// func TestDiscoveryHandlerNoDiscovery(t *testing.T) {
// 	defer setupSQLite()
//
// 	// delete discovery from table
// 	s := getSQLiteDB()
// 	defer s.disconnect()
//
// 	_, err := s.db.Exec("delete from taxii_discovery")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// now try to use handler
// 	ts := getStorer()
// 	defer ts.disconnect()
//
// 	req := httptest.NewRequest("GET", discoveryURL, nil)
// 	res := httptest.NewRecorder()
// 	h := handleTaxiiDiscovery(ts, testConfig().Port)
// 	h(res, req)
//
// 	if res.Code != http.StatusNotFound {
// 		t.Error("Got:", res.Code, "Expected:", http.StatusNotFound)
// 	}
// }
//
// func TestDiscoveryHandlerError(t *testing.T) {
// 	defer setupSQLite()
//
// 	// drop the table all together
// 	s := getSQLiteDB()
// 	defer s.disconnect()
//
// 	_, err := s.db.Exec("drop table taxii_discovery")
// 	if err != nil {
// 		t.Fatal(err)
// 	}
//
// 	// now try to use handler
// 	ts := getStorer()
// 	defer ts.disconnect()
//
// 	req := httptest.NewRequest("GET", discoveryURL, nil)
// 	res := httptest.NewRecorder()
// 	h := handleTaxiiDiscovery(ts, testConfig().Port)
// 	h(res, req)
//
// 	if res.Code != http.StatusNotFound {
// 		t.Error("Got:", res.Code, "Expected:", http.StatusNotFound)
// 	}
// }

/* helpers */

func TestInsertPort(t *testing.T) {
	tests := []struct {
		url      string
		port     int
		expected string
	}{
		{"http://test.com/foo", 1234, "http://test.com:1234/foo"},
		{"http://test.com", 1234, "http://test.com:1234"},
	}

	for _, test := range tests {
		result := insertPort(test.url, test.port)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestSwapPath(t *testing.T) {
	tests := []struct {
		url      string
		path     string
		expected string
	}{
		{"http://test.com/foo", "baz", "http://test.com/baz"},
		{"http://test.com", "foo", "http://test.com/foo"},
	}

	for _, test := range tests {
		result := swapPath(test.url, test.path)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestURLTokens(t *testing.T) {
	tests := []struct {
		url      string
		expected map[string]string
	}{
		{"http://test.com/foo", map[string]string{"scheme": "http", "host": "test.com", "path": "/foo"}},
		{"https://test.com", map[string]string{"scheme": "https", "host": "test.com", "path": ""}},
	}

	for _, test := range tests {
		result := urlTokens(test.url)

		if result["scheme"] != test.expected["scheme"] {
			t.Error("Got:", result["scheme"], "Expected:", test.expected["scheme"])
		}
		if result["host"] != test.expected["host"] {
			t.Error("Got:", result["host"], "Expected:", test.expected["host"])
		}
		if result["path"] != test.expected["path"] {
			t.Error("Got:", result["path"], "Expected:", test.expected["path"])
		}
	}
}
