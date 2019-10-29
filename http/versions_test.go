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

func TestVersionsHandlerGetPage(t *testing.T) {
	tests := []struct {
		limit    int
		expected int
	}{
		{1, 1},
		{10, 10},
	}

	for _, test := range tests {
		// set up mock service
		ms := mockVersionsService()
		ms.VersionsFn = func(ctx context.Context, cid, oid string, p *cabby.Page, f cabby.Filter) (cabby.Versions, error) {
			v := cabby.Versions{}
			for i := 0; i < test.expected; i++ {
				v.Versions = append(v.Versions, "")
			}

			p.Total = uint64(test.expected)
			return v, nil
		}
		h := VersionsHandler{VersionsService: ms}

		// set up request
		req := newRequest(http.MethodGet, testVersionsURL+"?limit="+strconv.Itoa(test.limit), nil)
		res := httptest.NewRecorder()

		h.Get(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		var result cabby.Versions
		err := json.Unmarshal([]byte(body), &result)
		if err != nil {
			t.Fatal(err)
		}

		if res.Code != http.StatusOK {
			t.Error("Got:", res.Code, "Expected:", http.StatusOK)
		}

		if len(result.Versions) != test.expected {
			t.Error("Got:", len(result.Versions), "Expected:", test.expected)
		}
	}
}

func TestVersionsHandlerGetInvalidPage(t *testing.T) {
	expected := cabby.Error{
		Title: "Bad Request", Description: "Invalid limit specified", HTTPStatus: http.StatusBadRequest}

	h := VersionsHandler{VersionsService: mockVersionsService()}
	status, body := handlerTest(h.Get, http.MethodGet, testVersionsURL+"?limit=0", nil)

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

func TestVersionsHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Version failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockVersionsService()
	ms.VersionsFn = func(ctx context.Context, cid, oid string, p *cabby.Page, f cabby.Filter) (cabby.Versions, error) {
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
	ms.VersionsFn = func(ctx context.Context, cid, oid string, p *cabby.Page, f cabby.Filter) (cabby.Versions, error) {
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
