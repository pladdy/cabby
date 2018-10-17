package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	"github.com/pladdy/stones"
)

func TestBundleFromBytesUnmarshalFail(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

func TestBundleFromBytesInvalidBundle(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"}`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

func TestGreaterThan(t *testing.T) {
	tests := []struct {
		x, y   int
		result bool
	}{
		{1, 2, false},
		{1, 1, false},
		{2, 1, true},
		{0, -1, true},
	}

	for _, test := range tests {
		if result := greaterThan(int64(test.x), int64(test.y)); result != test.result {
			t.Error("Got:", result, "Expected:", test.result)
		}
	}
}

func TestObjectsHandlerGet(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}

	// call handler for object
	req := newRequest("GET", testObjectURL, nil)
	req.Header.Set("Accept", cabby.StixContentType)
	status, body, _ := callHandler(h.Get, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	expected := tester.Object
	// objects are returned as bundles; collection ids are not defined in the bundle
	expected.CollectionID = cabby.ID{}
	expected.Object = nil

	// parse the bundle for an object
	var bundle stones.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	var object cabby.Object
	err = json.Unmarshal(bundle.Objects[0], &object)
	if err != nil {
		t.Fatal(err)
	}

	passed := tester.CompareObject(object, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsHandlerGetHeaders(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newRequest("GET", testObjectsURL, nil)
	req.Header.Set("Accept", cabby.StixContentType)

	res := httptest.NewRecorder()
	h.Get(res, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	tm := time.Time{}

	if res.Header().Get("Content-Type") != cabby.StixContentType {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", cabby.StixContentType)
	}
	if res.Header().Get("X-Taxii-Date-Added-First") != tm.Format(time.RFC3339Nano) {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", tm.Format(time.RFC3339Nano))
	}
	if res.Header().Get("X-Taxii-Date-Added-Last") != tm.Format(time.RFC3339Nano) {
		t.Error("Got:", res.Header().Get("Content-Type"), "Expected:", tm.Format(time.RFC3339Nano))
	}
}

func TestObjectsHandlerGetUnsupportedMimeType(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}

	// call handler for object
	req := newRequest("GET", testObjectURL, nil)
	req.Header.Set("Accept", "invalid")

	res := httptest.NewRecorder()
	h.Get(res, req)

	if res.Code != http.StatusUnsupportedMediaType {
		t.Error("Got:", res.Code, "Expected:", http.StatusUnsupportedMediaType)
	}
}

func TestObjectsHandlerGetObjectFailure(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Object failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
		return []cabby.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObject, "GET", testObjectURL, nil)

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

func TestObjectsHandlerGetObjectNoObject(t *testing.T) {
	s := mockObjectService()
	s.ObjectFn = func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
		return []cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObject, "GET", testObjectURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "No objects defined in this collection"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsHandlerGetObjects(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var bundle stones.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) <= 0 {
		t.Error("Got:", len(bundle.Objects), "Expected: > 0")
	}
}

func TestObjectsHandlerGetObjectsRange(t *testing.T) {
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
		obs := mockObjectService()
		obs.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
			objects := []cabby.Object{}
			for i := 0; i < test.expected; i++ {
				objects = append(objects, cabby.Object{})
			}

			cr.Total = uint64(test.expected)
			return objects, nil
		}
		h := ObjectsHandler{ObjectService: obs}

		// set up request
		req := newRequest("GET", testObjectsURL, nil)
		req.Header.Set("Accept", cabby.StixContentType)
		req.Header.Set("Range", "items "+strconv.Itoa(test.first)+"-"+strconv.Itoa(test.last))

		res := httptest.NewRecorder()
		h.Get(res, req)

		body, _ := ioutil.ReadAll(res.Body)

		var result stones.Bundle
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

func TestObjectsHandlerGetInvalidRange(t *testing.T) {
	tests := []struct {
		rangeString    string
		expectedStatus int
	}{
		{"items invalid", http.StatusRequestedRangeNotSatisfiable},
		{"items 0-0", http.StatusPartialContent},
	}

	h := ObjectsHandler{ObjectService: mockObjectService()}

	for _, test := range tests {
		// set up request
		req := newRequest("GET", testObjectsURL, nil)
		req.Header.Set("Accept", cabby.StixContentType)
		req.Header.Set("Range", test.rangeString)

		res := httptest.NewRecorder()
		h.Get(res, req)

		if res.Code != test.expectedStatus {
			t.Error("Got:", res.Code, "Expected:", test.expectedStatus)
		}
	}
}

func TestObjectsGetObjectsFailure(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
		return []cabby.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

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

func TestObjectsHandlerGetObjectsNoObjects(t *testing.T) {
	s := mockObjectService()
	s.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
		return []cabby.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObjects, "GET", testObjectsURL, nil)

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

/* Post */

func TestObjectsHandlerPost(t *testing.T) {
	osv := mockObjectService()
	osv.CreateBundleFn = func(ctx context.Context, b stones.Bundle, collectionID string, s cabby.Status, ss cabby.StatusService) {
		tester.Info.Println("mock call of CreateBundle")
	}

	ssv := mockStatusService()
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: osv, StatusService: ssv}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectsURL, b)
	status, body, headers := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusAccepted {
		t.Error("Got:", status, "Expected:", http.StatusAccepted)
	}

	if headers.Get("content-type") != cabby.TaxiiContentType {
		t.Error("Got:", headers["content-type"], "Expected:", cabby.TaxiiContentType)
	}

	var result cabby.Status
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Status != "pending" {
		t.Error("Got:", result.Status, "Expected: pending")
	}
	if result.PendingCount != 3 {
		t.Error("Got:", result.PendingCount, "Expected: 3")
	}
}

func TestObjectsHandlerPostForbidden(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}

	req := newPostRequest(testObjectsURL, nil)
	status, _, _ := callHandler(h.Post, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectsHandlerPostContentTooLarge(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(1), ObjectService: mockObjectService()}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestObjectsHandlerPostInvalidBundle(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	bundle := []byte(`{"foo":"bar"}`)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsHandlerPostEmptyBundle(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	// make a valid bundle except that the objects are mpty
	bundle := []byte(`{"type": "bundle", "objects": [], "spec_version": "2.0", "id": "bundle--5d0092c5-5f74-4287-9642-33f4c354e56d"}`)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsPostStatusFail(t *testing.T) {
	s := mockStatusService()
	s.CreateStatusFn = func(ctx context.Context, status cabby.Status) error { return errors.New("fail") }

	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService(), StatusService: &s}

	expected := cabby.Error{
		Title:       "Internal Server Error",
		Description: "Unable to store status resource",
		HTTPStatus:  http.StatusInternalServerError}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectsURL, b)
	status, body, headers := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != expected.HTTPStatus {
		t.Error("Got:", status, "Expected:", expected.HTTPStatus)
	}

	if headers.Get("content-type") != cabby.TaxiiContentType {
		t.Error("Got:", headers["content-type"], "Expected:", cabby.TaxiiContentType)
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

func TestObjectsHandlerPostToObjectURL(t *testing.T) {
	osv := mockObjectService()
	osv.CreateBundleFn = func(ctx context.Context, b stones.Bundle, collectionID string, s cabby.Status, ss cabby.StatusService) {
		tester.Info.Println("mock call of CreateBundle")
	}

	ssv := mockStatusService()
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: osv, StatusService: ssv}

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)
	b := bytes.NewBuffer(bundle)

	req := newPostRequest(testObjectURL, b)
	status, _, headers := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}

	if headers.Get("allow") != "Get, Head" {
		t.Error("Got:", headers["content-type"], "Expected:", "Get, Head")
	}
}

func TestObjectsPostValidPost(t *testing.T) {
	tests := []struct {
		accept      string
		contentType string
		valid       bool
	}{
		{cabby.TaxiiContentType, cabby.StixContentType, true},
		{"invalid", cabby.StixContentType, false},
		{cabby.TaxiiContentType, "invalid", false},
	}

	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	for _, test := range tests {
		r := newRequest("POST", testObjectsURL, nil)
		r.Header.Set("Accept", test.accept)
		r.Header.Set("Content-Type", test.contentType)

		w := httptest.NewRecorder()

		result := h.validPost(w, r.WithContext(cabby.WithUser(r.Context(), tester.User)))
		if result != test.valid {
			t.Error("Got:", result, "Expected:", test.valid)
		}
	}
}

func TestObjectsToBundleError(t *testing.T) {
	_, err := objectsToBundle([]cabby.Object{})
	if err == nil {
		t.Error("Expected error")
	}
}
