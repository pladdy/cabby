package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestDiscoveryHandlerHandleDiscovery(t *testing.T) {
	ds := DiscoveryService{}
	ds.ReadFn = func() (cabby.Result, error) {
		return cabby.Result{Data: testDiscovery()}, nil
	}

	// call handler
	h := DiscoveryHandler{DiscoveryService: &ds}
	status, result := handlerTest(h.HandleDiscovery(testPort), "GET", testDiscoveryURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.Discovery
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := testDiscovery()

	if discovery.Title != expected.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != expected.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
	if discovery.Contact != expected.Contact {
		t.Error("Got:", discovery.Contact, "Expected:", expected.Contact)
	}

	expectedDefault := insertPort(expected.Default, testPort)
	if discovery.Default != expectedDefault {
		t.Error("Got:", discovery.Default, "Expected:", expectedDefault)
	}
}

func TestDiscoveryHandlerHandleDiscoveryFailures(t *testing.T) {
	tests := []struct {
		method         string
		hasReadErr     bool
		errorDesc      string
		expectedError  cabby.Error
		expectedStatus int
	}{
		{method: "INVALID",
			hasReadErr:     true,
			errorDesc:      "Invalid method: INVALID",
			expectedError:  cabby.Error{Title: "Method Not Allowed"},
			expectedStatus: http.StatusMethodNotAllowed},
		{method: "GET",
			hasReadErr:     true,
			errorDesc:      "Read failure",
			expectedError:  cabby.Error{Title: "Internal Server Error"},
			expectedStatus: http.StatusInternalServerError},
		{method: "GET",
			hasReadErr:     false,
			errorDesc:      "Invalid result",
			expectedError:  cabby.Error{Title: "Internal Server Error"},
			expectedStatus: http.StatusInternalServerError},
	}

	for _, test := range tests {
		// finish setting up test
		test.expectedError.Description = test.errorDesc
		test.expectedError.HTTPStatus = test.expectedStatus

		ds := DiscoveryService{}
		ds.ReadFn = func() (cabby.Result, error) {
			var err error
			if test.hasReadErr {
				err = errors.New(test.errorDesc)
			}
			return cabby.Result{}, err
		}

		h := DiscoveryHandler{DiscoveryService: &ds}
		status, result := handlerTest(h.HandleDiscovery(testPort), test.method, testDiscoveryURL, nil)

		if status != test.expectedStatus {
			t.Error("Got:", status, "Expected:", test.expectedStatus)
		}

		var cabbyError cabby.Error
		err := json.Unmarshal([]byte(result), &cabbyError)
		if err != nil {
			t.Fatal(err)
		}

		if cabbyError.Title != test.expectedError.Title {
			t.Error("Got:", cabbyError.Title, "Expected:", test.expectedError.Title)
		}
		if cabbyError.Description != test.expectedError.Description {
			t.Error("Got:", cabbyError.Description, "Expected:", test.expectedError.Description)
		}
		if cabbyError.HTTPStatus != test.expectedError.HTTPStatus {
			t.Error("Got:", cabbyError.HTTPStatus, "Expected:", test.expectedError.HTTPStatus)
		}
	}
}

func TestDiscoveryHandlerNoDiscovery(t *testing.T) {
	ds := DiscoveryService{}
	ds.ReadFn = func() (cabby.Result, error) {
		return cabby.Result{Data: cabby.Discovery{Title: ""}}, nil
	}

	h := DiscoveryHandler{DiscoveryService: &ds}
	status, result := handlerTest(h.HandleDiscovery(testPort), "GET", testDiscoveryURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var cabbyError cabby.Error
	err := json.Unmarshal([]byte(result), &cabbyError)
	if err != nil {
		t.Fatal(err)
	}
	expected := cabby.Error{Title: "Resource not found",
		Description: "Discovery not defined", HTTPStatus: http.StatusNotFound}

	if cabbyError.Title != expected.Title {
		t.Error("Got:", cabbyError.Title, "Expected:", expected.Title)
	}
	if cabbyError.Description != expected.Description {
		t.Error("Got:", cabbyError.Description, "Expected:", expected.Description)
	}
	if cabbyError.HTTPStatus != expected.HTTPStatus {
		t.Error("Got:", cabbyError.HTTPStatus, "Expected:", expected.HTTPStatus)
	}
}

func TestDiscoveryHandlerAPIRoots(t *testing.T) {
	ds := DiscoveryService{}
	ds.ReadFn = func() (cabby.Result, error) {
		return cabby.Result{Data: testDiscovery()}, nil
	}

	h := DiscoveryHandler{DiscoveryService: &ds}
	status, result := handlerTest(h.HandleDiscovery(testPort), "GET", testDiscoveryURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.Discovery
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := testDiscovery()

	if discovery.Title != expected.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != expected.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
	if discovery.Contact != expected.Contact {
		t.Error("Got:", discovery.Contact, "Expected:", expected.Contact)
	}

	expectedDefault := insertPort(expected.Default, testPort)
	if discovery.Default != expectedDefault {
		t.Error("Got:", discovery.Default, "Expected:", expectedDefault)
	}

	expectedAPIRoots := []string{"https://localhost:1234/test_api_root/"}
	if discovery.APIRoots[0] != expectedAPIRoots[0] {
		t.Error("Got:", discovery.APIRoots[0], "Expected:", expectedAPIRoots[0])
	}
}

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
