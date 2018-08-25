package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestBundleFromBytesUnmarshalFail(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

func TestBundleFromBytesInvalidBundle(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"}`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

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
	h := ObjectsHandler{ObjectService: mockObjectService()}
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

func TestObjectsHandlerGetObjectFailure(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return cabby.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObject, "GET", testObjectURL, nil)

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

func TestObjectsHandlerGetObjectNoObject(t *testing.T) {
	s := mockObjectService()
	s.ObjectFn = func(collectionID, objectID string) (cabby.Object, error) {
		return cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
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
	h := ObjectsHandler{ObjectService: mockObjectService()}
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

func TestObjectsGetObjectsFailure(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

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

func TestObjectsHandlerGetObjectsNoObjects(t *testing.T) {
	s := mockObjectService()
	s.ObjectsFn = func(collectionID string) ([]cabby.Object, error) {
		return []cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
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
	h := ObjectsHandler{
		MaxContentLength: int64(2048),
		ObjectService:    mockObjectService(),
		StatusService:    mockStatusService()}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	status, body := handlerTest(h.Post, "POST", testObjectURL, b)

	if status != http.StatusAccepted {
		t.Error("Got:", status, "Expected:", http.StatusAccepted)
	}

	var result cabby.Status
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "pending" {
		t.Error("Got:", result.Status, "Expected: pending")
	}
	if result.PendingCount != 3 {
		t.Error("Got:", result.PendingCount, "Expected: 3")
	}
}

func TestObjectsHandlerPostForbidden(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, _ := handlerTestNoAuth(h.Post, "POST", testObjectURL, nil)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectsHandlerPostContentTooLarge(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(1), ObjectService: mockObjectService()}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	status, _ := handlerTest(h.Post, "POST", testObjectsURL, b)

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestObjectsHandlerPostInvalidBundle(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	bundle := []byte(`{"foo":"bar"}`)
	b := bytes.NewBuffer(bundle)

	status, _ := handlerTest(h.Post, "POST", testObjectsURL, b)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsHandlerPostEmptyBundle(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	// make a valid bundle except that the objects are mpty
	bundle := []byte(`{"type": "bundle", "objects": [], "spec_version": "2.0", "id": "bundle--5d0092c5-5f74-4287-9642-33f4c354e56d"}`)
	b := bytes.NewBuffer(bundle)

	status, _ := handlerTest(h.Post, "POST", testObjectsURL, b)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsPostStatusFail(t *testing.T) {
	s := mockStatusService()
	s.CreateStatusFn = func(status cabby.Status) error { return errors.New("fail") }

	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService(), StatusService: &s}

	expected := cabby.Error{
		Title:       "Internal Server Error",
		Description: "Unable to store status resource",
		HTTPStatus:  http.StatusInternalServerError}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	status, body := handlerTest(h.Post, "POST", testObjectsURL, b)

	if status != expected.HTTPStatus {
		t.Error("Got:", status, "Expected:", expected.HTTPStatus)
	}

	fmt.Println(body)

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
