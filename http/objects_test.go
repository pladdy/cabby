package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestObjectsHandlerGet(t *testing.T) {
	os := tester.ObjectService{}
	os.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return tester.Objects, nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &os}
	status, result := handlerTest(h.Get, "GET", testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery []cabby.Object
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.Objects

	if len(expected) <= 0 {
		t.Error("Got:", len(expected), "Expected: > 0")
	}
}

func TestObjectsGetFailures(t *testing.T) {
	tests := []struct {
		method   string
		expected cabby.Error
	}{
		{method: "GET",
			expected: cabby.Error{
				Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}},
	}

	for _, test := range tests {
		expected := test.expected

		os := tester.ObjectService{}
		os.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
			return []cabby.Object{}, errors.New(expected.Description)
		}

		h := ObjectsHandler{ObjectService: &os}
		status, body := handlerTest(h.Get, test.method, testObjectsURL, nil)

		if status != expected.HTTPStatus {
			t.Error("Got:", status, "Expected:", expected.HTTPStatus)
		}

		var result cabby.Error
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		passed := tester.CompareError(result, expected)
		if !passed {
			t.Error("Comparison failed")
		}
	}
}

func TestObjectsHandlerNoObjects(t *testing.T) {
	os := tester.ObjectService{}
	os.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &os}
	status, body := handlerTest(h.Get, "GET", testObjectsURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := cabby.Error{Title: "Resource not found",
		Description: "No objects defined in this collection", HTTPStatus: http.StatusNotFound}

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}
