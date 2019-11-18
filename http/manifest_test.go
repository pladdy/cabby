package http

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestManifestHandlerDelete(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestManifestHandlerGet(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
	status, body, _ := callHandler(h.Get, req)

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

func TestManifestHandlerGetBadRequest(t *testing.T) {
	expected := cabby.Error{
		Title: "Bad Request", Description: "Invalid limit specified", HTTPStatus: http.StatusBadRequest}

	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newClientRequest(http.MethodGet, testManifestURL+"?limit=0", nil)
	status, body, _ := callHandler(h.Get, req)

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

func TestManifestHandlerGetForbidden(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestManifestHandlerGetHeaders(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
	res := httptest.NewRecorder()
	tm := time.Time{}
	h.Get(res, req)

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

func TestManifestHandlerGetNotAcceptable(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
	req.Header.Set("Accept", "invalid")
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusNotAcceptable {
		t.Error("Got:", status, "Expected:", http.StatusNotAcceptable)
	}
}

func TestManifestHandlerGetInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Manifest failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockManifestService()
	ms.ManifestFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, errors.New(expected.Description)
	}

	h := ManifestHandler{ManifestService: &ms}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
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

func TestManifestHandlerGetNoManifest(t *testing.T) {
	ms := mockManifestService()
	ms.ManifestFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) (cabby.Manifest, error) {
		return cabby.Manifest{}, nil
	}

	h := ManifestHandler{ManifestService: &ms}
	req := newClientRequest(http.MethodGet, testManifestURL, nil)
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
	expected.Description = "No resources available for this request"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestManifestHandlerGetPage(t *testing.T) {
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

		req := newClientRequest(http.MethodGet, testManifestURL+"?limit="+strconv.Itoa(test.limit), nil)
		status, body, _ := callHandler(h.Get, req)

		var result cabby.Manifest
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		if len(result.Objects) != test.expected {
			t.Error("Got:", len(result.Objects), "Expected:", test.expected)
		}
	}
}

func TestManifestHandlerPost(t *testing.T) {
	h := ManifestHandler{ManifestService: mockManifestService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
