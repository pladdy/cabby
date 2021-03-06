package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestCollectionsHandlerDelete(t *testing.T) {
	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testCollectionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestCollectionsHandlerGet(t *testing.T) {
	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, result := handlerTest(h.Get, http.MethodGet, testCollectionsURL, nil)

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

func TestCollectionsHandlerGetBadRequest(t *testing.T) {
	expected := cabby.Error{
		Title: "Bad Request", Description: "Invalid limit specified", HTTPStatus: http.StatusBadRequest}

	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, body := handlerTest(h.Get, http.MethodGet, testCollectionsURL+"?limit=0", nil)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
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

func TestCollectionsHandlerGetInternalSeverError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	cs := mockCollectionService()
	cs.CollectionsFn = func(ctx context.Context, apiRootPath string, p *cabby.Page) (cabby.Collections, error) {
		return cabby.Collections{}, errors.New(expected.Description)
	}

	h := CollectionsHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, http.MethodGet, testCollectionsURL, nil)

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
	cs.CollectionsFn = func(ctx context.Context, apiRoot string, p *cabby.Page) (cabby.Collections, error) {
		return cabby.Collections{}, nil
	}

	h := CollectionsHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, http.MethodGet, testCollectionsURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No resources available for this request"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionsHandlerGetCollectionsNonExistant(t *testing.T) {
	nonRoutedCollection, _ := cabby.NewID()
	badRoute := testCollectionsURL + nonRoutedCollection.String() + "/"

	cs := mockCollectionService()
	h := CollectionsHandler{CollectionService: &cs}
	status, body := handlerTest(h.Get, http.MethodGet, badRoute, nil)

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
		t.Error("Comparison failed, result:", result, "Expected:", expected)
	}
}

func TestCollectionsHandlerGetPage(t *testing.T) {
	tests := []struct {
		limit    int
		expected int
	}{
		{1, 1},
		{10, 10},
	}

	for _, test := range tests {
		// set up mock service
		cs := mockCollectionService()
		cs.CollectionsFn = func(ctx context.Context, apiRootPath string, p *cabby.Page) (cabby.Collections, error) {
			collections := cabby.Collections{}
			for i := 0; i < test.expected; i++ {
				collections.Collections = append(collections.Collections, cabby.Collection{})
			}

			p.Total = uint64(test.expected)
			return collections, nil
		}
		h := CollectionsHandler{CollectionService: cs}

		// set up request
		req := newClientRequest(http.MethodGet, testCollectionsURL+"?limit="+strconv.Itoa(test.limit), nil)
		status, body, _ := callHandler(h.Get, req)

		var result cabby.Collections
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		if len(result.Collections) != test.expected {
			t.Error("Got:", len(result.Collections), "Expected:", test.expected)
		}
	}
}

func TestCollectionsHandlerPost(t *testing.T) {
	h := CollectionsHandler{CollectionService: mockCollectionService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testCollectionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
