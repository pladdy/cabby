package http

import (
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestManifestHandleDelete(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestManifestHandlerGet(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, body := handlerTest(h.Get, http.MethodGet, testManifestURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Manifest
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.ManifestEntry

	passed := tester.CompareManifestEntry(result.Objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestManifestHandlerGetHeaders(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newRequest(http.MethodGet, testManifestURL, nil)

	res := httptest.NewRecorder()
	h.Get(res, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	tm := time.Time{}

	if res.Header().Get("Content-Type") != cabby.TaxiiContentType {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", cabby.TaxiiContentType)
	}
	if res.Header().Get("X-Taxii-Date-Added-First") != tm.Format(time.RFC3339Nano) {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", tm.Format(time.RFC3339Nano))
	}
	if res.Header().Get("X-Taxii-Date-Added-Last") != tm.Format(time.RFC3339Nano) {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", tm.Format(time.RFC3339Nano))
	}
}

func TestManifestHandlerGetRange(t *testing.T) {
	tests := []struct {
		limit    int
		expected int
	}{
		{1, 1},
		{10, 10},
	}

	for _, test := range tests {
		// set up mock service
		ms := mockManifestService()
		ms.ManifestFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) (cabby.Manifest, error) {
			manifest := cabby.Manifest{}
			for i := 0; i < test.expected; i++ {
				manifest.Objects = append(manifest.Objects, cabby.ManifestEntry{})
			}

			p.Total = uint64(test.expected)
			return manifest, nil
		}
		h := ManifestHandler{ManifestService: ms}

		// set up request
		req := newRequest(http.MethodGet, testManifestURL+"?limit="+strconv.Itoa(test.limit), nil)
		res := httptest.NewRecorder()

		h.Get(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		var result cabby.Manifest
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if res.Code != http.StatusOK {
			t.Error("Got:", res.Code, "Expected:", http.StatusOK)
		}

		if len(result.Objects) != test.expected {
			t.Error("Got:", len(result.Objects), "Expected:", test.expected)
		}
	}
}

func TestManifestHandlerGetInvalidPage(t *testing.T) {
	expected := cabby.Error{
		Title: "Bad Request", Description: "Invalid limit specified", HTTPStatus: http.StatusBadRequest}

	h := ManifestHandler{ManifestService: mockManifestService()}
	status, body := handlerTest(h.Get, http.MethodGet, testManifestURL+"?limit=0", nil)

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

func TestManifestHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Manifest failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockManifestService()
	ms.ManifestFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, errors.New(expected.Description)
	}

	h := ManifestHandler{ManifestService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testManifestURL, nil)

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

func TestManifestHandlerGetNoManifest(t *testing.T) {
	ms := mockManifestService()
	ms.ManifestFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, nil
	}

	h := ManifestHandler{ManifestService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testManifestURL, nil)

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

func TestManifestHandlePost(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
