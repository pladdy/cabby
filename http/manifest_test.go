package http

import (
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestManifestHandlerGet(t *testing.T) {
	ms := tester.ManifestService{}
	ms.ManifestFn = func(collectionID string) (cabby.Manifest, error) {
		return tester.Manifest, nil
	}

	// call handler
	h := ManifestHandler{ManifestService: &ms}
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

func TestManifestHandlerGetFailures(t *testing.T) {
	tests := []struct {
		method   string
		expected cabby.Error
	}{
		{method: "GET",
			expected: cabby.Error{
				Title: "Internal Server Error", Description: "Manifest failure", HTTPStatus: http.StatusInternalServerError}},
	}

	for _, test := range tests {
		expected := test.expected

		ms := tester.ManifestService{}
		ms.ManifestFn = func(collectionID string) (cabby.Manifest, error) {
			return cabby.Manifest{}, errors.New(expected.Description)
		}

		h := ManifestHandler{ManifestService: &ms}
		status, body := handlerTest(h.Get, test.method, testManifestURL, nil)

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
}

func TestManifestHandlerGetNoManifest(t *testing.T) {
	ms := tester.ManifestService{}
	ms.ManifestFn = func(collectionID string) (cabby.Manifest, error) {
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
	expected.Description = "No manifest available for this collection"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestManifestHandlePost(t *testing.T) {
	ms := tester.ManifestService{}
	ms.ManifestFn = func(collectionID string) (cabby.Manifest, error) {
		return cabby.Manifest{}, nil
	}

	h := ManifestHandler{ManifestService: &ms}
	status, _ := handlerTest(h.Post, "POST", testManifestURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}
