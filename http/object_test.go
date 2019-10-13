package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	"github.com/pladdy/stones"
)

/* Delete */

func TestObjectsHandleDelete(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testObjectsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestObjectsHandlerDelete(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testObjectURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}
}

func TestObjectsHandlerDeleteObjectBadRequest(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.DeleteObjectFn = func(ctx context.Context, collectionID, objectID string) error {
		return errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.Delete, http.MethodDelete, testObjectURL, nil)

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

/* Get */

func TestObjectsHandlerGet(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}

	// call handler for object
	req := newRequest(http.MethodGet, testObjectURL, nil)
	req.Header.Set("Accept", cabby.TaxiiContentType)
	status, body, _ := callHandler(h.Get, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	expected := tester.Object

	// parse the envelope for an object
	var envelope cabby.Envelope
	err := json.Unmarshal([]byte(body), &envelope)
	if err != nil {
		t.Fatal(err)
	}

	var object stones.Object
	err = json.Unmarshal(envelope.Objects[0], &object)
	if err != nil {
		t.Fatal(err)
	}

	passed := tester.CompareObject(object, expected)
	if !passed {
		t.Error("Comparison failed", "\nObject:", object, "\nExpected:", expected)
	}
}

func TestObjectsHandlerGetObjectFailure(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObject, http.MethodGet, testObjectURL, nil)

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
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObject, http.MethodGet, testObjectURL, nil)

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
