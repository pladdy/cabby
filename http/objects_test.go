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

func TestObjectsHandlerDelete(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	status, _ := handlerTest(h.Delete, http.MethodDelete, testObjectsURL, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Got:", status, "Expected:", http.StatusMethodNotAllowed)
	}
}

/* Get */

func TestObjectsHandlerGet(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodGet, testObjectsURL, nil)
	status, body, _ := callHandler(h.Get, req)

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

func TestObjectsHandlerGetBadRequest(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodGet, testObjectsURL+"?limit=0", nil)
	status, body, _ := callHandler(h.Get, req)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := cabby.Error{
		Title: "Bad Request", Description: "Invalid limit specified", HTTPStatus: http.StatusBadRequest}

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsHandlerGetHeaders(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodGet, testObjectsURL, nil)
	res := httptest.NewRecorder()
	h.Get(res, req)

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

func TestObjectsHandlerGetForbidden(t *testing.T) {
	h := ObjectsHandler{ObjectService: mockObjectService()}
	req := newClientRequest(http.MethodGet, testObjectsURL, nil)
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Get, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectsHandlerGetInternalServerError(t *testing.T) {
	expected := cabby.Error{
		Title: "Internal Server Error", Description: "Collection failure", HTTPStatus: http.StatusInternalServerError}

	s := mockObjectService()
	s.ObjectsFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, errors.New(expected.Description)
	}

	h := ObjectsHandler{ObjectService: &s}
	req := newClientRequest(http.MethodGet, testObjectsURL, nil)
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

func TestObjectsHandlerGetNoObjects(t *testing.T) {
	s := mockObjectService()
	s.ObjectsFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) ([]stones.Object, error) {
		return []stones.Object{}, nil
	}

	h := ObjectsHandler{ObjectService: &s}
	req := newClientRequest(http.MethodGet, testObjectsURL, nil)
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

func TestObjectsHandlerGetPage(t *testing.T) {
	tests := []struct {
		limit    int
		expected int
	}{
		{1, 1},
		{10, 10},
	}

	for _, test := range tests {
		// set up mock service
		obs := mockObjectService()
		obs.ObjectsFn = func(ctx context.Context, collectionID string, p *cabby.Page, f cabby.Filter) ([]stones.Object, error) {
			objects := []stones.Object{}
			for i := 0; i < test.expected; i++ {
				objects = append(objects, tester.GenerateObject("malware"))
			}

			p.Total = uint64(test.expected)
			return objects, nil
		}
		h := ObjectsHandler{ObjectService: obs}

		req := newClientRequest(http.MethodGet, testObjectsURL+"?limit="+strconv.Itoa(test.limit), nil)
		status, body, _ := callHandler(h.Get, req)

		var result cabby.Envelope
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

	req := newClientRequest(http.MethodPost, testObjectsURL, bytes.NewBuffer(envelope))
	status, body, headers := callHandler(h.Post, req)

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

func TestObjectsHandlerPostContentTooLarge(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(1), ObjectService: mockObjectService()}

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)
	b := bytes.NewBuffer(envelope)

	req := newClientRequest(http.MethodPost, testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req)

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestObjectsHandlerPostEmptyEnvelope(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	// make a valid envelope except that the objects are empty
	envelope := []byte(`{"objects": []}`)
	b := bytes.NewBuffer(envelope)

	req := newClientRequest(http.MethodPost, testObjectsURL, b)
	status, _, _ := callHandler(h.Post, req)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestObjectsHandlerPostForbidden(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: 8192, ObjectService: mockObjectService()}

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	envelope, _ := ioutil.ReadAll(envelopeFile)

	req := newClientRequest(http.MethodPost, testObjectsURL, bytes.NewBuffer(envelope))
	req = req.WithContext(context.Background())
	status, _, _ := callHandler(h.Post, req)

	if status != http.StatusForbidden {
		t.Error("Got:", status, "Expected:", http.StatusForbidden)
	}
}

func TestObjectsHandlerPostInvalidEnvelope(t *testing.T) {
	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	envelope := []byte(`{"foo":"bar"`)
	req := newClientRequest(http.MethodPost, testObjectsURL, bytes.NewBuffer(envelope))
	status, _, _ := callHandler(h.Post, req)

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

	req := newClientRequest(http.MethodPost, testObjectsURL, bytes.NewBuffer(envelope))
	status, body, headers := callHandler(h.Post, req)

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

func TestValidPost(t *testing.T) {
	tests := []struct {
		accept      string
		contentType string
		valid       bool
	}{
		{cabby.TaxiiContentType, cabby.TaxiiContentType, true},
		{cabby.TaxiiContentType, "invalid", false},
	}

	h := ObjectsHandler{MaxContentLength: int64(2048), ObjectService: mockObjectService()}

	for _, test := range tests {
		r := newClientRequest(http.MethodPost, testObjectsURL, nil)
		r.Header.Set("Accept", test.accept)
		r.Header.Set("Content-Type", test.contentType)
		w := httptest.NewRecorder()

		result := h.validPost(w, r)
		if result != test.valid {
			t.Error("Got:", result, "Expected:", test.valid)
		}
	}
}
