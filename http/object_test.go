package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestObjectHandlerGet(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return tester.Object, nil
	}

	// call handler
	h := ObjectHandler{ObjectService: &ds}
	status, body := handlerTest(h.Get, "GET", testObjectURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Object
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Object

	passed := tester.CompareObject(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectGetFailures(t *testing.T) {
	tests := []struct {
		method   string
		expected cabby.Error
	}{
		{method: "GET",
			expected: cabby.Error{
				Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}},
	}

	for _, test := range tests {
		expected := test.expected

		ds := tester.ObjectService{}
		ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
			return cabby.Object{}, errors.New(expected.Description)
		}

		h := ObjectHandler{ObjectService: &ds}
		status, body := handlerTest(h.Get, test.method, testObjectURL, nil)

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

func TestObjectHandlerNoObject(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return cabby.Object{}, nil
	}

	h := ObjectHandler{ObjectService: &ds}
	status, body := handlerTest(h.Get, "GET", testObjectURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := cabby.Error{Title: "Resource not found",
		Description: "Object ID doesn't exist in this collection", HTTPStatus: http.StatusNotFound}

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}
