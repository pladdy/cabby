package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestCollectionHandlerGet(t *testing.T) {
	ds := tester.CollectionService{}
	ds.CollectionFn = func(user, apiRootPath, collectionID string) (cabby.Collection, error) {
		return tester.Collection, nil
	}

	// call handler
	h := CollectionHandler{CollectionService: &ds}
	status, body := handlerTest(h.Get, "GET", testCollectionURL, nil)

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

		ds := tester.CollectionService{}
		ds.CollectionFn = func(user, apiRootPath, collectionID string) (cabby.Collection, error) {
			return cabby.Collection{}, errors.New(expected.Description)
		}

		h := CollectionHandler{CollectionService: &ds}
		status, body := handlerTest(h.Get, test.method, testCollectionURL, nil)

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

func TestCollectionHandlerGetNoCollection(t *testing.T) {
	ds := tester.CollectionService{}
	ds.CollectionFn = func(user, apiRootPath, collectionID string) (cabby.Collection, error) {
		return cabby.Collection{}, nil
	}

	h := CollectionHandler{CollectionService: &ds}
	status, body := handlerTest(h.Get, "GET", testCollectionURL, nil)

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

func TestCollectionHandlePost(t *testing.T) {
	ds := tester.CollectionService{}
	ds.CollectionFn = func(user, apiRootPath, collectionID string) (cabby.Collection, error) {
		return cabby.Collection{}, nil
	}

	h := CollectionHandler{CollectionService: &ds}
	status, _ := handlerTest(h.Post, "POST", testCollectionURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
