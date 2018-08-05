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
	"regexp"
	"strings"
	"testing"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
)

func TestRequireAcceptType(t *testing.T) {
	tests := []struct {
		acceptedHeader string
		acceptHeader   string
		responseCode   int
	}{
		{StixContentType, "application/vnd.oasis.stix+json; version=2.0", http.StatusOK},
		{StixContentType, "application/vnd.oasis.stix+json", http.StatusOK},
		{StixContentType, "application/vnd.oasis.stix+json;verion=2.0", http.StatusOK},
		{StixContentType, "", http.StatusUnsupportedMediaType},
		{StixContentType, "application/vnd.oasis.stix+jsonp", http.StatusUnsupportedMediaType},
		{StixContentType, "application/vnd.oasis.stix+jsonp; version=3.0", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+json; version=2.0", http.StatusOK},
		{TaxiiContentType, "application/vnd.oasis.taxii+json", http.StatusOK},
		{TaxiiContentType, "application/vnd.oasis.taxii+json;verion=2.0", http.StatusOK},
		{TaxiiContentType, "", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+jsonp", http.StatusUnsupportedMediaType},
		{TaxiiContentType, "application/vnd.oasis.taxii+jsonp; version=3.0", http.StatusUnsupportedMediaType},
	}

	for _, test := range tests {
		mockHandler := func(w http.ResponseWriter, r *http.Request) {
			accept := r.Header.Get("Accept")
			io.WriteString(w, fmt.Sprintf("Accept Header: %v", accept))
		}
		mockHandler = WithAcceptType(mockHandler, test.acceptedHeader)

		req := httptest.NewRequest("GET", "/test", nil)
		req.Header.Add("Accept", test.acceptHeader)
		res := httptest.NewRecorder()

		mockHandler(res, req)
		body, _ := ioutil.ReadAll(res.Body)

		if res.Code != test.responseCode {
			t.Error("Got:", res.Code, string(body), "Expected:", http.StatusOK)
		}
	}
}

func TestWithBasicAuth(t *testing.T) {
	tests := []struct {
		userFn         func(user, password string) (cabby.User, error)
		validFn        func(cabby.User) bool
		expectedStatus int
	}{
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{Email: testUserEmail}, nil
		},
			validFn:        func(cabby.User) bool { return true },
			expectedStatus: http.StatusOK},
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{}, errors.New("service error")
		},
			validFn:        func(cabby.User) bool { return true },
			expectedStatus: http.StatusInternalServerError},
		{userFn: func(user, password string) (cabby.User, error) {
			return cabby.User{}, nil
		},
			validFn:        func(cabby.User) bool { return false },
			expectedStatus: http.StatusUnauthorized},
	}

	for _, test := range tests {
		// set up service
		us := UserService{}
		us.UserFn = test.userFn
		us.ValidFn = test.validFn

		// set up handler
		mockHandler := mockHandler(t.Name())
		decoratedHandler := withBasicAuth(mockHandler, &us)

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
	mockHandler := mockHandler(t.Name())
	decoratedHandler := withRequestLogging(mockHandler)

	// set up a server
	server := httptest.NewServer(decoratedHandler)
	defer server.Close()

	req := newServerRequest("GET", server.URL)
	res, _ := getResponse(req, server)

	if res.StatusCode != http.StatusOK {
		t.Error("Got:", res.StatusCode, "Expected:", http.StatusOK)
	}

	// get last log
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	lastLog := logs[len(logs)-1]

	type expectedLog struct {
		Time      string
		Level     string
		Msg       string
		ElapsedTs float64 `json:"elapsed_ts"`
		EndTs     int64   `json:"end_ts"`
		Method    string
		URL       string
	}

	// parse log into struct
	var result expectedLog
	err := json.Unmarshal([]byte(lastLog), &result)
	if err != nil {
		t.Fatal(err)
	}

	if len(result.Time) <= 0 {
		t.Error("Got:", result.Time, "Expected: a time")
	}
	if result.Level != "info" {
		t.Error("Got:", result.Level, "Expected: info")
	}
	if len(result.Msg) <= 0 {
		t.Error("Got:", result.Msg, "Expected: a message")
	}
	if result.ElapsedTs < 0 {
		t.Error("Got:", result.ElapsedTs, "Expected: elapsed time >= 0 ms")
	}
	if result.EndTs < 0 {
		t.Error("Got:", result.EndTs, "Expected: end time > 0 ms")
	}
	if len(result.Method) <= 0 {
		t.Error("Got:", result.Msg, "Expected: a method")
	}
	if len(result.URL) <= 0 {
		t.Error("Got:", result.Msg, "Expected: a URL")
	}
}

/* helpers */

func getResponse(req *http.Request, server *httptest.Server) (*http.Response, error) {
	c := server.Client()
	return c.Do(req)
}

func mockHandler(testName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testName)
	})
}

func newServerRequest(method, url string) *http.Request {
	req := httptest.NewRequest("GET", url, nil)
	// this can't be set in client requests
	req.RequestURI = ""
	// the values don't matter, but have to be set for
	req.SetBasicAuth("user", "password")
	return req
}
