package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestAPIRootHandlerGet(t *testing.T) {
	ds := tester.APIRootService{}
	ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
		return tester.APIRoot, nil
	}

	// call handler
	h := APIRootHandler{APIRootService: &ds}
	status, result := handlerTest(h.Get, "GET", testAPIRootURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.APIRoot
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.APIRoot

	if discovery.Title != expected.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != expected.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
}

func TestAPIRootGetFailures(t *testing.T) {
	tests := []struct {
		method   string
		expected cabby.Error
	}{
		{method: "GET",
			expected: cabby.Error{
				Title: "Internal Server Error", Description: "APIRoot failure", HTTPStatus: http.StatusInternalServerError}},
	}

	for _, test := range tests {
		expected := test.expected

		ds := tester.APIRootService{}
		ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
			return cabby.APIRoot{}, errors.New(expected.Description)
		}

		h := APIRootHandler{APIRootService: &ds}
		status, body := handlerTest(h.Get, test.method, testAPIRootURL, nil)

		if status != expected.HTTPStatus {
			t.Error("Got:", status, "Expected:", expected.HTTPStatus)
		}

		var result cabby.Error
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		tester.CompareError(result, expected, t)
	}
}

func TestAPIRootHandlerNoAPIRoot(t *testing.T) {
	ds := tester.APIRootService{}
	ds.APIRootFn = func(path string) (cabby.APIRoot, error) {
		return cabby.APIRoot{Title: ""}, nil
	}

	h := APIRootHandler{APIRootService: &ds}
	status, result := handlerTest(h.Get, "GET", testAPIRootURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var cabbyError cabby.Error
	err := json.Unmarshal([]byte(result), &cabbyError)
	if err != nil {
		t.Fatal(err)
	}
	expected := cabby.Error{Title: "Resource not found",
		Description: "API Root not found: cabby_test_root", HTTPStatus: http.StatusNotFound}

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
