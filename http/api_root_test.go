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

func TestAPIRootHandleDelete(t *testing.T) {
	as := mockAPIRootService()
	h := APIRootHandler{APIRootService: &as}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testAPIRootURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestAPIRootHandlerGet(t *testing.T) {
	h := APIRootHandler{APIRootService: mockAPIRootService()}
	status, body := handlerTest(h.Get, http.MethodGet, testAPIRootURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.APIRoot
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.APIRoot

	passed := tester.CompareAPIRoot(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestAPIRootHandlerGetFailures(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "APIRoot failure", HTTPStatus: http.StatusInternalServerError}

	as := mockAPIRootService()
	as.APIRootFn = func(ctx context.Context, path string) (cabby.APIRoot, error) {
		return cabby.APIRoot{}, errors.New(expected.Description)
	}

	h := APIRootHandler{APIRootService: &as}
	status, body := handlerTest(h.Get, http.MethodGet, testAPIRootURL, nil)

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

func TestAPIRootHandlerGetNoAPIRoot(t *testing.T) {
	as := mockAPIRootService()
	as.APIRootFn = func(ctx context.Context, path string) (cabby.APIRoot, error) {
		return cabby.APIRoot{}, nil
	}

	h := APIRootHandler{APIRootService: &as}
	req := newClientRequest(http.MethodGet, testAPIRootURL, nil)
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
	expected.Description = "API Root not found"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestAPIRootHandlerGetNotAcceptable(t *testing.T) {
	h := APIRootHandler{APIRootService: mockAPIRootService()}
	req := newClientRequest(http.MethodGet, testAPIRootURL, nil)
	req.Header.Set("Accept", "invalid")
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusNotAcceptable {
		t.Error("Got:", status, "Expected:", http.StatusNotAcceptable)
	}
}

func TestAPIRootHandlePost(t *testing.T) {
	h := APIRootHandler{APIRootService: mockAPIRootService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testAPIRootURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
