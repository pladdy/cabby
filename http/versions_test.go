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

func TestVersionHandleDelete(t *testing.T) {
	h := VersionsHandler{VersionsService: mockVersionsService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testVersionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestVersionsHandlerGet(t *testing.T) {
	h := VersionsHandler{VersionsService: mockVersionsService()}
	status, body := handlerTest(h.Get, http.MethodGet, testVersionsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Versions
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := "2016-04-06T20:07:09.000Z"
	if result.Versions[0] != expected {
		t.Error("Got:", result.Versions[0], "Expected:", expected)
	}
}

func TestVersionsHandlerGetHeaders(t *testing.T) {
	h := VersionsHandler{VersionsService: mockVersionsService()}
	req := newRequest(http.MethodGet, testVersionsURL, nil)

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

func TestVersionsHandlerGetRange(t *testing.T) {
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
		ms := mockVersionsService()
		ms.VersionsFn = func(ctx context.Context, cid, oid string, cr *cabby.Range, f cabby.Filter) (cabby.Versions, error) {
			v := cabby.Versions{}
			for i := 0; i < test.expected; i++ {
				v.Versions = append(v.Versions, "")
			}

			cr.Total = uint64(test.expected)
			return v, nil
		}
		h := VersionsHandler{VersionsService: ms}

		// set up request
		req := newRequest(http.MethodGet, testVersionsURL, nil)
		req.Header.Set("Range", "items "+strconv.Itoa(test.first)+"-"+strconv.Itoa(test.last))

		res := httptest.NewRecorder()
		h.Get(res, req)

		body, _ := ioutil.ReadAll(res.Body)

		var result cabby.Versions
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if res.Code != http.StatusPartialContent {
			t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
		}

		if len(result.Versions) != test.expected {
			t.Error("Got:", len(result.Versions), "Expected:", test.expected)
		}

		ra := cabby.Range{First: uint64(test.first), Last: uint64(test.last), Total: uint64(test.expected)}
		if res.Header().Get("Content-Range") != ra.String() {
			t.Error("Got:", res.Header().Get("Content-Range"), "Expected:", ra.String())
		}
	}
}

func TestVersionsHandlerGetInvalidRange(t *testing.T) {
	tests := []struct {
		rangeString    string
		expectedStatus int
	}{
		{"items invalid", http.StatusRequestedRangeNotSatisfiable},
		{"items 0-0", http.StatusPartialContent},
	}

	h := VersionsHandler{VersionsService: mockVersionsService()}

	for _, test := range tests {
		// set up request
		req := newRequest(http.MethodGet, testVersionsURL, nil)
		req.Header.Set("Range", test.rangeString)

		res := httptest.NewRecorder()
		h.Get(res, req)

		if res.Code != test.expectedStatus {
			t.Error("Got:", res.Code, "Expected:", test.expectedStatus)
		}
	}
}

func TestVersionsHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Version failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockVersionsService()
	ms.VersionsFn = func(ctx context.Context, cid, oid string, cr *cabby.Range, f cabby.Filter) (cabby.Versions, error) {
		return cabby.Versions{}, errors.New(expected.Description)
	}

	h := VersionsHandler{VersionsService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testVersionsURL, nil)

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

func TestVersionsHandlerGetNoVersion(t *testing.T) {
	ms := mockVersionsService()
	ms.VersionsFn = func(ctx context.Context, cid, oid string, cr *cabby.Range, f cabby.Filter) (cabby.Versions, error) {
		return cabby.Versions{}, nil
	}

	h := VersionsHandler{VersionsService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testVersionsURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err, result, body)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No resources available for this request"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestVersionHandlePost(t *testing.T) {
	h := VersionsHandler{VersionsService: mockVersionsService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testVersionsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
