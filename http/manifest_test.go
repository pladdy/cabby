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

func TestManifestHandlerGet(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, body := handlerTest(h.Get, "GET", testManifestURL, nil)

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
	req := newRequest("GET", testManifestURL, nil)

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
		first    int
		last     int
		expected int
	}{
		{0, 0, 1},
		{0, 9, 10},
	}

	for _, test := range tests {
		// set up mock service
		ms := mockManifestService()
		ms.ManifestFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
			manifest := cabby.Manifest{}
			for i := 0; i < test.expected; i++ {
				manifest.Objects = append(manifest.Objects, cabby.ManifestEntry{})
			}

			cr.Total = uint64(test.expected)
			return manifest, nil
		}
		h := ManifestHandler{ManifestService: ms}

		// set up request
		req := newRequest("GET", testManifestURL, nil)
		req.Header.Set("Range", "items "+strconv.Itoa(test.first)+"-"+strconv.Itoa(test.last))

		res := httptest.NewRecorder()
		h.Get(res, req)

		body, _ := ioutil.ReadAll(res.Body)

		var result cabby.Manifest
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if res.Code != http.StatusPartialContent {
			t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
		}

		if len(result.Objects) != test.expected {
			t.Error("Got:", len(result.Objects), "Expected:", test.expected)
		}

		ra := cabby.Range{First: uint64(test.first), Last: uint64(test.last), Total: uint64(test.expected)}
		if res.Header().Get("Content-Range") != ra.String() {
			t.Error("Got:", res.Header().Get("Content-Range"), "Expected:", ra.String())
		}
	}
}

func TestManifestHandlerGetInvalidRange(t *testing.T) {
	tests := []struct {
		rangeString    string
		expectedStatus int
	}{
		{"items invalid", http.StatusRequestedRangeNotSatisfiable},
		{"items 0-0", http.StatusPartialContent},
	}

	h := ManifestHandler{ManifestService: mockManifestService()}

	for _, test := range tests {
		// set up request
		req := newRequest("GET", testManifestURL, nil)
		req.Header.Set("Range", test.rangeString)

		res := httptest.NewRecorder()
		h.Get(res, req)

		if res.Code != test.expectedStatus {
			t.Error("Got:", res.Code, "Expected:", test.expectedStatus)
		}
	}
}

func TestManifestHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Manifest failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockManifestService()
	ms.ManifestFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, errors.New(expected.Description)
	}

	h := ManifestHandler{ManifestService: &ms}
	status, body := handlerTest(h.Get, "GET", testManifestURL, nil)

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
	ms.ManifestFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, nil
	}

	h := ManifestHandler{ManifestService: &ms}
	status, body := handlerTest(h.Get, "GET", testManifestURL, nil)

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
	status, _ := handlerTest(h.Post, "POST", testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
