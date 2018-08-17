package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
	log "github.com/sirupsen/logrus"
)

func TestHandleUndefinedRequest(t *testing.T) {
	status, body := handlerTest(handleUndefinedRoute, "GET", "/", nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}

	expected := cabby.Error{Title: "Resource not found", Description: "Invalid path: /", HTTPStatus: 404}

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

// set up mock handler
type mockRequestHandler struct {
}

func (m mockRequestHandler) Get(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		io.WriteString(w, "valid http method")
	default:
		io.WriteString(w, "invalid http method")
	}
}

func TestRequestHandlerRouteRequest(t *testing.T) {
	mock := mockRequestHandler{}

	tests := []struct {
		method string
		url    string
		status int
	}{
		{"CUSTOM", testDiscoveryURL, http.StatusMethodNotAllowed},
		{"GET", testDiscoveryURL, http.StatusOK},
	}

	for _, test := range tests {
		status, _ := handlerTest(RouteRequest(mock), test.method, test.url, nil)

		if status != test.status {
			t.Error("Got:", status, "Expected:", test.status)
		}
	}
}

func TestWithAcceptType(t *testing.T) {
	tests := []struct {
		acceptedHeader string
		acceptHeader   string
		responseCode   int
	}{
		{cabby.StixContentType, "application/vnd.oasis.stix+json; version=2.0", http.StatusOK},
		{cabby.StixContentType, "application/vnd.oasis.stix+json", http.StatusOK},
		{cabby.StixContentType, "application/vnd.oasis.stix+json;verion=2.0", http.StatusOK},
		{cabby.StixContentType, "", http.StatusUnsupportedMediaType},
		{cabby.StixContentType, "application/vnd.oasis.stix+jsonp", http.StatusUnsupportedMediaType},
		{cabby.StixContentType, "application/vnd.oasis.stix+jsonp; version=3.0", http.StatusUnsupportedMediaType},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json; version=2.0", http.StatusOK},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json", http.StatusOK},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+json;verion=2.0", http.StatusOK},
		{cabby.TaxiiContentType, "", http.StatusUnsupportedMediaType},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+jsonp", http.StatusUnsupportedMediaType},
		{cabby.TaxiiContentType, "application/vnd.oasis.taxii+jsonp; version=3.0", http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		testHandler := func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")
			io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
		}
		testHandler = WithAcceptType(testHandler, test.acceptedHeader)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		testHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", http.StatusOK)
		}
	}
}

func TestWithBasicAuth(t *testing.T) {
	tests := []struct {
		userFn         func(user, password string) (cabby.User, error)
		existsFn       func(cabby.User) bool
		expectedStatus int
	}{
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{Email: tester.UserEmail}, nil
		},
			existsFn:       func(cabby.User) bool { return true },
			expectedStatus: http.StatusOK},
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{}, errors.New("service error")
		},
			existsFn:       func(cabby.User) bool { return true },
			expectedStatus: http.StatusInternalServerError},
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{}, nil
		},
			existsFn:       func(cabby.User) bool { return false },
			expectedStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		// set up service
		us := tester.UserService{}
		us.UserFn = test.userFn
		us.ExistsFn = test.existsFn

		// set up handler
		testHandler := testHandler(t.Name())
		decoratedHandler := withBasicAuth(testHandler, &us)

		// set up a server
		server := httptest.NewServer(decoratedHandler)
		defer server.Close()

		req := newServerRequest("GET", server.URL)
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
	decoratedHandler := withRequestLogging(testHandler)

	// set up a server
	server := httptest.NewServer(decoratedHandler)
	defer server.Close()

	req := newServerRequest("GET", server.URL)
	res, _ := getResponse(req, server)

	if res.StatusCode != http.StatusOK {
		t.Error("Got:", res.StatusCode, "Expected:", http.StatusOK)
	}

	// parse log into struct
	var result requestLog
	err := json.Unmarshal([]byte(lastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	testRequestLog(result, t)
}
