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

func TestObjectHandlerDelete(t *testing.T) {
	h := ObjectHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodDelete, testObjectURL, nil)
	status, _, _ := callHandler(h.Delete, req)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}
}

func TestObjectHandlerDeleteForbidden(t *testing.T) {
	h := ObjectHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodDelete, testObjectURL, nil)
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Delete, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectHandlerDeleteInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.DeleteObjectFn = func(ctx context.Context, collectionID, objectID string) error {
		return errors.New(expected.Description)
	}

	h := ObjectHandler{ObjectService: &s}
	req := newClientRequest(http.MethodDelete, testObjectURL, nil)
	status, body, _ := callHandler(h.Delete, req)

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

func TestObjectHandlerGet(t *testing.T) {
	h := ObjectHandler{ObjectService: mockObjectService()}

	// call handler for object
	req := newClientRequest(http.MethodGet, testObjectURL, nil)
	status, body, _ := callHandler(h.Get, req)

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

func TestObjectHandlerGetForbidden(t *testing.T) {
	h := ObjectHandler{ObjectService: mockObjectService()}

	req := newClientRequest(http.MethodGet, testObjectURL, nil)
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectHandlerGetInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, errors.New(expected.Description)
	}

	h := ObjectHandler{ObjectService: &s}
	req := newClientRequest(http.MethodGet, testObjectURL, nil)
	status, body, _ := callHandler(h.Get, req)

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

func TestObjectHandlerGetNoObject(t *testing.T) {
	s := mockObjectService()
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, nil
	}

	h := ObjectHandler{ObjectService: &s}
	req := newClientRequest(http.MethodGet, testObjectURL, nil)
	status, body, _ := callHandler(h.Get, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No objects found"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed; result:", result, "Expected:", expected)
	}
}

func TestObjectHandlerPost(t *testing.T) {
	h := ObjectHandler{ObjectService: mockObjectService()}
	status, _ := handlerTest(h.Post, http.MethodDelete, testObjectsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
