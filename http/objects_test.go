package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		x, y   int
		result bool
	}{
		{1, 2, false},
		{1, 1, false},
		{2, 1, true},
		{0, -1, true},
	}

	for _, test := range tests {
		if result := greaterThan(int64(test.x), int64(test.y)); result != test.result {
			t.Error("Got:", result, "Expected:", test.result)
		}
	}
}

func TestObjectsHandlerGet(t *testing.T) {
	ds := tester.ObjectService{}
	ds.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return tester.Object, nil
	}
	ds.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{tester.Object}, nil
	}

	h := ObjectsHandler{ObjectService: &ds}
	expected := tester.Object

	// call handler for object
	status, body := handlerTest(h.Get, "GET", testObjectURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var object cabby.Object
	err := json.Unmarshal([]byte(body), &object)
	if err != nil {
		t.Fatal(err)
	}

	passed := tester.CompareObject(object, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// call handler for objects
	status, body = handlerTest(h.Get, "GET", testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var objects []cabby.Object
	err = json.Unmarshal([]byte(body), &objects)
	if err != nil {
		t.Fatal(err)
	}

	passed = tester.CompareObject(objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsHandlerGetObjectFailures(t *testing.T) {
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

	expected := tester.ErrorResourceNotFound
	expected.Description = "Object ID doesn't exist in this collection"

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
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result []cabby.Object
	err := json.Unmarshal([]byte(body), &result)
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

	expected := tester.ErrorResourceNotFound
	expected.Description = "No objects defined in this collection"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

/* Post */

func TestObjectsHandlerPost(t *testing.T) {
	ds := tester.ObjectService{}
	ds.CreateObjectFn = func(object cabby.Object) error {
		return nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &ds}
	status, _ := handlerTest(h.Post, "POST", testObjectURL, nil)

	if status != http.StatusAccepted {
		t.Error("Got:", status, "Expected:", http.StatusAccepted)
	}

	// var result cabby.Object
	// err := json.Unmarshal([]byte(body), &result)
	// if err != nil {
	// 	t.Fatal(err)
	// }
	//
	// expected := tester.Object
	//
	// passed := tester.CompareObject(result, expected)
	// if !passed {
	// 	t.Error("Comparison failed")
	// }
}

func TestObjectsHandlerPostForbidden(t *testing.T) {
	ds := tester.ObjectService{}
	ds.CreateObjectFn = func(object cabby.Object) error {
		return nil
	}

	// call handler
	h := ObjectsHandler{ObjectService: &ds}
	status, _ := handlerTestNoAuth(h.Post, "POST", testObjectURL, nil)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectsHandlerPostContentTooLarge(t *testing.T) {
	ds := tester.ObjectService{MaxContentLength: int64(1)}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	b := bytes.NewBuffer(bundle)
	h := ObjectsHandler{ObjectService: &ds}
	status, _ := handlerTest(h.Post, "POST", testObjectsURL, b)

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}
