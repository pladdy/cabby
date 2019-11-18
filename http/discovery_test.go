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

func TestDiscoveryHandlerDelete(t *testing.T) {
	ds := mockDiscoveryService()
	ds.DiscoveryFn = func(ctx context.Context) (cabby.Discovery, error) {
		return cabby.Discovery{Title: ""}, nil
	}

	h := DiscoveryHandler{DiscoveryService: &ds, Port: tester.Port}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testDiscoveryURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestDiscoveryHandlerGet(t *testing.T) {
	h := DiscoveryHandler{DiscoveryService: mockDiscoveryService(), Port: tester.Port}
	status, body := handlerTest(h.Get, http.MethodGet, testDiscoveryURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var result cabby.Discovery
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}
	expected := tester.Discovery

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestDiscoveryHandlerGetInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Discovery failure", HTTPStatus: http.StatusInternalServerError}

	ds := mockDiscoveryService()
	ds.DiscoveryFn = func(ctx context.Context) (cabby.Discovery, error) {
		return cabby.Discovery{}, errors.New(expected.Description)
	}

	h := DiscoveryHandler{DiscoveryService: &ds, Port: tester.Port}
	status, body := handlerTest(h.Get, http.MethodGet, testDiscoveryURL, nil)

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

func TestDiscoveryHandlerGetNoDiscovery(t *testing.T) {
	ds := mockDiscoveryService()
	ds.DiscoveryFn = func(ctx context.Context) (cabby.Discovery, error) {
		return cabby.Discovery{Title: ""}, nil
	}

	h := DiscoveryHandler{DiscoveryService: &ds, Port: tester.Port}
	status, body := handlerTest(h.Get, http.MethodGet, testDiscoveryURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "Discovery not defined"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestDiscoveryHandlerPost(t *testing.T) {
	ds := mockDiscoveryService()
	ds.DiscoveryFn = func(ctx context.Context) (cabby.Discovery, error) {
		return cabby.Discovery{Title: ""}, nil
	}

	h := DiscoveryHandler{DiscoveryService: &ds, Port: tester.Port}
	status, _ := handlerTest(h.Post, http.MethodPost, testDiscoveryURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestInsertPort(t *testing.T) {
	tests := []struct {
		url      string
		port     int
		expected string
	}{
		{"http://test.com/foo", 1234, "http://test.com:1234/foo"},
		{"http://test.com", 1234, "http://test.com:1234"},
		{"http://test.com/foo/", 1234, "http://test.com:1234/foo/"},
	}

	for _, test := range tests {
		result := insertPort(test.url, test.port)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

func TestSwapPath(t *testing.T) {
	tests := []struct {
		url      string
		path     string
		expected string
	}{
		{"http://test.com/foo", "baz", "http://test.com/baz"},
		{"http://test.com", "foo", "http://test.com/foo"},
		{"http://test.com/foo/", "baz/", "http://test.com/baz/"},
		{"http://test.com/", "foo/", "http://test.com/foo/"},
	}

	for _, test := range tests {
		result := swapPath(test.url, test.path)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}
