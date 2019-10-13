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
	log "github.com/sirupsen/logrus"
)

func TestEnvelopeFromBytes(t *testing.T) {
	envelopeFile, err := os.Open("testdata/malware_envelope.json")
	if err != nil {
		t.Fatal(err)
	}

	rawEnvelope, err := ioutil.ReadAll(envelopeFile)
	if err != nil {
		t.Fatal(err)
	}

	_, err = envelopeFromBytes(rawEnvelope)
	if err != nil {
		t.Error("Error unexpected:", err)
	}
}

func TestEnvelopeFromBytesUnmarshalFail(t *testing.T) {
	b, err := envelopeFromBytes([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for envelope:", b)
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

/* Get */

func TestObjectsHandlerGetHeaders(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newRequest(http.MethodGet, testObjectsURL, nil)
	req.Header.Set("Accept", cabby.TaxiiContentType)

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

func TestObjectsHandlerGetObjects(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, body := handlerTest(h.getObjects, http.MethodGet, testObjectsURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var envelope cabby.Envelope
	err := json.Unmarshal([]byte(body), &envelope)
	if err != nil {
		t.Fatal(err)
	}

	if len(envelope.Objects) <= 0 {
		t.Error("Got:", len(envelope.Objects), "Expected: > 0")
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
		obs.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
			objects := []stones.Object{}
			for i := 0; i < test.expected; i++ {
				objects = append(objects, tester.GenerateObject("malware"))
			}

			cr.Total = uint64(test.expected)
			return objects, nil
		}
		h := ObjectsHandler{ObjectService: obs}

		// set up request
		req := newRequest(http.MethodGet, testObjectsURL, nil)
		req.Header.Set("Accept", cabby.TaxiiContentType)
		req.Header.Set("Range", "items "+strconv.Itoa(test.first)+"-"+strconv.Itoa(test.last))

		res := httptest.NewRecorder()
		h.Get(res, req)

		body, _ := ioutil.ReadAll(res.Body)

		var result cabby.Envelope
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
		req := newRequest(http.MethodGet, testObjectsURL, nil)
		req.Header.Set("Accept", cabby.TaxiiContentType)
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
	s.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObjects, http.MethodGet, testObjectsURL, nil)

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
	s.ObjectsFn = func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
	status, body := handlerTest(h.getObjects, http.MethodGet, testObjectsURL, nil)

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
	osv.CreateEnvelopeFn = func(ctx context.Context, e cabby.Envelope, collectionID string, s cabby.Status, ss cabby.StatusService) {
		log.Debug("mock call of CreateEnvelope")
	}

	ssv := mockStatusService()
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: osv, StatusService: ssv}

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)
	b := bytes.NewBuffer(envelope)

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

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)
	b := bytes.NewBuffer(envelope)

	req := newPostRequest(testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestObjectsHandlerPostInvalidEnvelope(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	envelope := []byte(`{"foo":"bar"`)
	b := bytes.NewBuffer(envelope)

	req := newPostRequest(testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsHandlerPostEmptyEnvelope(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	// make a valid envelope except that the objects are empty
	envelope := []byte(`{"objects": []}`)
	b := bytes.NewBuffer(envelope)

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

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)
	b := bytes.NewBuffer(envelope)

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

func TestObjectsHandlePostToObjectURL(t *testing.T) {
	osv := mockObjectService()
	osv.CreateEnvelopeFn = func(ctx context.Context, e cabby.Envelope, collectionID string, s cabby.Status, ss cabby.StatusService) {
		log.Debug("mock call of CreateEnvelope")
	}

	ssv := mockStatusService()
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: osv, StatusService: ssv}

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)
	b := bytes.NewBuffer(envelope)

	req := newPostRequest(testObjectURL, b)
	status, _, _ := callHandler(h.Post, req.WithContext(cabby.WithUser(req.Context(), tester.User)))

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestObjectsPostValidPost(t *testing.T) {
	tests := []struct {
		accept      string
		contentType string
		valid       bool
	}{
		{cabby.TaxiiContentType, cabby.TaxiiContentType, true},
		{"invalid", cabby.TaxiiContentType, false},
		{cabby.TaxiiContentType, "invalid", false},
	}

	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	for _, test := range tests {
		r := newRequest(http.MethodPost, testObjectsURL, nil)
		r.Header.Set("Accept", test.accept)
		r.Header.Set("Content-Type", test.contentType)

		w := httptest.NewRecorder()

		result := h.validPost(w, r.WithContext(cabby.WithUser(r.Context(), tester.User)))
		if result != test.valid {
			t.Error("Got:", result, "Expected:", test.valid)
		}
	}
}

func TestObjectsToEnvelopeError(t *testing.T) {
	_, err := objectsToEnvelope([]stones.Object{})
	if err == nil {
		t.Error("Expected error")
	}
}
