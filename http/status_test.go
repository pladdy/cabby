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

func TestStatusHandlerDelete(t *testing.T) {
	h := StatusHandler{StatusService: mockStatusService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testStatusURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestStatusHandlerGet(t *testing.T) {
	h := StatusHandler{StatusService: mockStatusService()}
	status, body := handlerTest(h.Get, http.MethodGet, testStatusURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Status
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.Status

	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestStatusHandlerGetInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Status failure", HTTPStatus: http.StatusInternalServerError}

	ms := mockStatusService()
	ms.StatusFn = func(ctx context.Context, statusID string) (cabby.Status, error) {
		return cabby.Status{}, errors.New(expected.Description)
	}

	h := StatusHandler{StatusService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testStatusURL, nil)

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

func TestStatusHandlerGetNoStatus(t *testing.T) {
	ms := mockStatusService()
	ms.StatusFn = func(ctx context.Context, statusID string) (cabby.Status, error) {
		return cabby.Status{}, nil
	}

	h := StatusHandler{StatusService: &ms}
	status, body := handlerTest(h.Get, http.MethodGet, testStatusURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No status available for this id"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestStatusHandlePost(t *testing.T) {
	h := StatusHandler{StatusService: mockStatusService()}
	status, _ := handlerTest(h.Post, http.MethodPost, testStatusURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
