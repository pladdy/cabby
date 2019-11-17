package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestCollectionHandleDelete(t *testing.T) {
	h := CollectionHandler{CollectionService: mockCollectionService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testCollectionURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestCollectionHandlerGet(t *testing.T) {
	h := CollectionHandler{CollectionService: mockCollectionService()}
	req := newClientRequest(http.MethodGet, testCollectionURL, nil)
	status, body, _ := callHandler(h.Get, req)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Collection
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.Collection

	passed := tester.CompareCollection(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	cs := mockCollectionService()
	cs.CollectionFn = func(ctx context.Context, apiRootPath, collectionID string) (cabby.Collection, error) {
		return cabby.Collection{}, errors.New(expected.Description)
	}

	h := CollectionHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, http.MethodGet, testCollectionURL, nil)

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

func TestCollectionHandlerGetForbidden(t *testing.T) {
	h := CollectionHandler{CollectionService: mockCollectionService()}
	req := newClientRequest(http.MethodGet, testCollectionURL, nil)
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestCollectionHandlerGetNoCollection(t *testing.T) {
	cs := mockCollectionService()
	cs.CollectionFn = func(ctx context.Context, apiRootPath, collectionID string) (cabby.Collection, error) {
		return cabby.Collection{}, nil
	}

	h := CollectionHandler{CollectionService: &cs}
	req := newClientRequest(http.MethodGet, testCollectionURL, nil)
	status, body, _ := callHandler(h.Get, req)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "Collection ID doesn't exist in this API Root"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionHandlerGetNotAcceptable(t *testing.T) {
	h := CollectionHandler{CollectionService: mockCollectionService()}
	req := newClientRequest(http.MethodGet, testCollectionURL, nil)
	req.Header.Set("Accept", "invalid")
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusNotAcceptable {
		t.Error("Got:", status, "Expected:", http.StatusNotAcceptable)
	}
}

func TestCollectionHandlePost(t *testing.T) {
	h := CollectionHandler{CollectionService: mockCollectionService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testCollectionURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
