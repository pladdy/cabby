package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestObjectsHandlerObject(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return tester.Object, nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &ds}
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

func TestObjectsHandlerObjects(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{tester.Object}, nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &ds}
	status, body := handlerTest(h.Get, "GET", testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result []cabby.Object
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Object

	passed := tester.CompareObject(result[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsHandlerGetObject(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return tester.Object, nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &ds}
	status, body := handlerTest(h.getObject, "GET", testObjectURL, nil)

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

func TestObjectsGetObjectFailures(t *testing.T) {
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

		h := ObjectsHandler{ObjectService: &ds}
		status, body := handlerTest(h.getObject, test.method, testObjectURL, nil)

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

func TestObjectsHandlerGetObjectNoObject(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &ds}
	status, body := handlerTest(h.getObject, "GET", testObjectURL, nil)

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

func TestObjectsHandlerGetObjects(t *testing.T) {
	os := tester.ObjectService{}
	os.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return tester.Objects, nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &os}
	status, result := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

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

func TestObjectsGetObjectsFailures(t *testing.T) {
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
		status, body := handlerTest(h.getObjects, test.method, testObjectsURL, nil)

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

func TestObjectsHandlerGetObjectsNoObjects(t *testing.T) {
	os := tester.ObjectService{}
	os.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &os}
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

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
