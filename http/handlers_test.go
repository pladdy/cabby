package http

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	log "github.com/sirupsen/logrus"
)

func TestHandleUndefinedRequest(t *testing.T) {
	status, body := handlerTest(handleUndefinedRoute, http.MethodGet, "/", nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	var result cabby.Error
	err := json.Unmarshal([]byte(body), &result)
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.ErrorResourceNotFound
	expected.Description = "Invalid path: /"

	passed := tester.CompareError(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestWithAcceptSet(t *testing.T) {
	tests := []struct {
		acceptedHeader string
		acceptHeader   string
		responseCode   int
	}{
		{cabby.StixContentType, "application/vnd.oasis.stix+json; version=2.0", http.StatusOK},
		{cabby.StixContentType, "application/vnd.oasis.stix+json", http.StatusOK},
		{cabby.StixContentType, "application/vnd.oasis.stix+json;verion=2.0", http.StatusOK},
		{cabby.StixContentType, "", http.StatusNotAcceptable},
		{cabby.StixContentType, "application/vnd.oasis.stix+jsonp", http.StatusNotAcceptable},
		{cabby.StixContentType, "application/vnd.oasis.stix+jsonp; version=3.0", http.StatusNotAcceptable},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json; version=2.0", http.StatusOK},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json", http.StatusOK},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json;verion=2.0", http.StatusOK},
		{cabby.TaxiiContentType, "", http.StatusNotAcceptable},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+jsonp", http.StatusNotAcceptable},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+jsonp; version=3.0", http.StatusNotAcceptable},
	}

	for _, test := range tests {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")
			io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
		}
		testHandler = WithAcceptSet(testHandler, test.acceptedHeader)

		req := httptest.NewRequest(http.MethodGet, "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		testHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", test.responseCode)
		}
	}
}

func TestWithBasicAuth(t *testing.T) {
	tests := []struct {
		expectedStatus    int
		userFn            func(ctx context.Context, user, password string) (cabby.User, error)
		userCollectionsFn func(ctx context.Context, user string) (cabby.UserCollectionList, error)
	}{
		{expectedStatus: http.StatusOK,
			userFn: func(ctx context.Context, user, password string) (cabby.User, error) {
				return cabby.User{Email: tester.UserEmail}, nil
			},
			userCollectionsFn: func(ctx context.Context, user string) (cabby.UserCollectionList, error) {
				return cabby.UserCollectionList{}, nil
			}},
		{expectedStatus: http.StatusInternalServerError,
			userFn: func(ctx context.Context, user, password string) (cabby.User, error) {
				return cabby.User{}, errors.New("service error")
			},
			userCollectionsFn: func(ctx context.Context, user string) (cabby.UserCollectionList, error) {
				return cabby.UserCollectionList{}, nil
			}},
		{expectedStatus: http.StatusUnauthorized,
			userFn: func(ctx context.Context, user, password string) (cabby.User, error) {
				return cabby.User{}, nil
			},
			userCollectionsFn: func(ctx context.Context, user string) (cabby.UserCollectionList, error) {
				return cabby.UserCollectionList{}, nil
			}},
		{expectedStatus: http.StatusInternalServerError,
			userFn: func(ctx context.Context, user, password string) (cabby.User, error) {
				return cabby.User{Email: tester.UserEmail}, nil
			},
			userCollectionsFn: func(ctx context.Context, user string) (cabby.UserCollectionList, error) {
				return cabby.UserCollectionList{}, errors.New("service error")
			}},
	}

	for _, test := range tests {
		// set up service
		us := tester.UserService{}
		us.UserFn = test.userFn
		us.UserCollectionsFn = test.userCollectionsFn

		// set up handler
		testHandler := testHandler(t.Name())
		decoratedHandler := withBasicAuth(testHandler, &us)

		// set up a server
		server := httptest.NewServer(decoratedHandler)
		defer server.Close()

		req := newServerRequest(http.MethodGet, server.URL)
		res, _ := getResponse(req, server)

		if res.StatusCode != test.expectedStatus {
			t.Error("Got:", res.StatusCode, "Expected:", test.expectedStatus)
		}
	}
}

func TestWithRequestLogging(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// set up handler
	testHandler := testHandlerFunc(t.Name())
	decoratedHandler := withLogging(testHandler)

	// set up a server
	server := httptest.NewServer(decoratedHandler)
	defer server.Close()

	req := newServerRequest(http.MethodGet, server.URL)
	res, _ := getResponse(req, server)

	if res.StatusCode != http.StatusOK {
		t.Error("Got:", res.StatusCode, "Expected:", http.StatusOK)
	}

	// parse log into struct
	var result tester.RequestLog
	err := json.Unmarshal([]byte(tester.LastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testRequestLog(result, t)
}
