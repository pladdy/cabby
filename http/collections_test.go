package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestCollectionsHandlerGet(t *testing.T) {
	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, result := handlerTest(h.Get, "GET", testCollectionsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var discovery cabby.Collections
	err := json.Unmarshal([]byte(result), &discovery)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.Collections

	if len(expected.Collections) <= 0 {
		t.Error("Got:", len(expected.Collections), "Expected: > 0")
	}
}

func TestCollectionsHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	cs := mockCollectionService()
	cs.CollectionsFn = func(user, apiRootPath string) (cabby.Collections, error) {
		return cabby.Collections{}, errors.New(expected.Description)
	}

	h := CollectionsHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, "GET", testCollectionsURL, nil)

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

func TestCollectionsHandlerGetNoCollections(t *testing.T) {
	cs := mockCollectionService()
	cs.CollectionsFn = func(user, apiRoot string) (cabby.Collections, error) {
		return cabby.Collections{}, nil
	}

	h := CollectionsHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, "GET", testCollectionsURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No collections defined in this API Root"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionsHandlePost(t *testing.T) {
	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, _ := handlerTest(h.Post, "POST", testCollectionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
