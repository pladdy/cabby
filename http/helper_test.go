package http

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

var (
	testAPIRootURL     = tester.BaseURL + tester.APIRootPath + "/"
	testCollectionsURL = testAPIRootURL + "/collections/"
	testCollectionURL  = testCollectionsURL + tester.CollectionID + "/"
	testObjectsURL     = testCollectionURL + "/objects/"
	testObjectURL      = testObjectsURL + tester.ObjectID + "/"
	testDiscoveryURL   = tester.BaseURL + "/taxii/"
)

type requestLog struct {
	Time      string
	Level     string
	Msg       string
	ElapsedTs float64 `json:"elapsed_ts"`
	EndTs     int64   `json:"end_ts"`
	Method    string
	URL       string
}

/* helper functions */

func attemptRequest(c *http.Client, r *http.Request) (*http.Response, error) {
	tester.Info.Println("Requesting", r.URL, "from test server")
	MaxTries := 3

	for i := 0; i < MaxTries; i++ {
		res, err := c.Do(r)
		if err == nil {
			return res, err
		}

		tester.Warn.Println("  Web server for test not responding, waiting...")
		time.Sleep(time.Duration(i+1) * time.Second)
	}

	return nil, errors.New("Failed to get request")
}

func getResponse(req *http.Request, server *httptest.Server) (*http.Response, error) {
	c := server.Client()
	return c.Do(req)
}

// test a HandlerFunc.  given a HandlerFunc, method, url, and bytes.Buffer, call the request and record response.
// it returns the status code and response as a string
func handlerTest(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	var req *http.Request

	if b != nil {
		req = withAuthentication(httptest.NewRequest(method, url, b))
	} else {
		req = withAuthentication(httptest.NewRequest(method, url, nil))
	}

	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

func lastLog(buf bytes.Buffer) string {
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	return logs[len(logs)-1]
}

func newServerRequest(method, url string) *http.Request {
	req := httptest.NewRequest("GET", url, nil)

	// this can't be set in client requests
	req.RequestURI = ""

	// the values don't matter, but have to be set in the request
	req.SetBasicAuth("user", "password")

	req.Header.Set("Accept", cabby.TaxiiContentType)
	return req
}

func testHandler(testName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testName)
	}
}

func testHandlerFunc(testName string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testName)
	})
}

func testRequestLog(result requestLog, t *testing.T) {
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

// set up a http client that uses TLS
func tlsClient() *http.Client {
	tlsConfig := &tls.Config{
		InsecureSkipVerify: true,
	}
	tr := &http.Transport{TLSClientConfig: tlsConfig}
	return &http.Client{Transport: tr}
}

// create a context for the testUser and give it read/write access to the test collection
func withAuthentication(r *http.Request) *http.Request {
	ctx := context.WithValue(context.Background(), userName, tester.UserEmail)
	ctx = context.WithValue(ctx, canAdmin, true)
	return r.WithContext(ctx)
}
