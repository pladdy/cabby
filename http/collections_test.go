package http

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
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

func TestCollectionsHandlerGetRange(t *testing.T) {
	tests := []struct {
		first    int
		last     int
		expected int
	}{
		{0, 0, 1},
		{0, 9, 10},
	}

	for _, test := range tests {
		// set up mock service
		cs := mockCollectionService()
		cs.CollectionsFn = func(user, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
			collections := cabby.Collections{}
			for i := 0; i < test.expected; i++ {
				collections.Collections = append(collections.Collections, cabby.Collection{})
			}

			cr.Total = int64(test.expected)
			return collections, nil
		}
		h := CollectionsHandler{CollectionService: cs}

		// set up request
		req := withAuthentication(newRequest("GET", testCollectionsURL, nil))
		req.Header.Set("Range", "items "+strconv.Itoa(test.first)+"-"+strconv.Itoa(test.last))

		res := httptest.NewRecorder()
		h.Get(res, req)

		body, _ := ioutil.ReadAll(res.Body)

		var collections cabby.Collections
		err := json.Unmarshal([]byte(body), &collections)
		if err != nil {
			t.Fatal(err)
		}

		if res.Code != http.StatusPartialContent {
			t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
		}

		if len(collections.Collections) != test.expected {
			t.Error("Got:", len(collections.Collections), "Expected:", test.expected)
		}

		ra := cabby.Range{First: int64(test.first), Last: int64(test.last), Total: int64(test.expected)}
		if res.Header().Get("Content-Range") != ra.String() {
			t.Error("Got:", res.Header().Get("Content-Range"), "Expected:", ra.String())
		}
	}
}

func TestCollectionsHandlerGetInvalidRange(t *testing.T) {
	tests := []struct {
		rangeString    string
		expectedStatus int
	}{
		{"items invalid", http.StatusRequestedRangeNotSatisfiable},
		{"items 0-0", http.StatusPartialContent},
	}

	h := CollectionsHandler{CollectionService: mockCollectionService()}

	for _, test := range tests {
		// set up request
		req := withAuthentication(newRequest("GET", testCollectionsURL, nil))
		req.Header.Set("Range", test.rangeString)

		res := httptest.NewRecorder()
		h.Get(res, req)

		if res.Code != test.expectedStatus {
			t.Error("Got:", res.Code, "Expected:", test.expectedStatus)
		}
	}
}

func TestCollectionsHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	cs := mockCollectionService()
	cs.CollectionsFn = func(user, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
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
	cs.CollectionsFn = func(user, apiRoot string, cr *cabby.Range) (cabby.Collections, error) {
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
