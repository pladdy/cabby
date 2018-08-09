package http

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"
	"time"

	cabby "github.com/pladdy/cabby2"
)

const (
	eightMB          = 8388608
	testPort         = 1234
	testAPIRootPath  = "cabby_test_root"
	testAPIRootURL   = "https://localhost:1234/" + testAPIRootPath + "/"
	testDiscoveryURL = "https://localhost:1234/taxii/"
	testUserEmail    = "test@cabby.com"
	testUserPassword = "test"
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

/* mock services */

type mockDataStore struct {
	APIRootServiceFn   func() APIRootService
	DiscoveryServiceFn func() DiscoveryService
	UserServiceFn      func() UserService
}

func newMockDataStore() *mockDataStore {
	return &mockDataStore{}
}

func (s mockDataStore) APIRootService() cabby.APIRootService {
	return s.APIRootServiceFn()
}

func (s mockDataStore) Close() {
	return
}

func (s mockDataStore) DiscoveryService() cabby.DiscoveryService {
	return s.DiscoveryServiceFn()
}

func (s mockDataStore) Open() error {
	return nil
}

func (s mockDataStore) UserService() cabby.UserService {
	return s.UserServiceFn()
}

type APIRootService struct {
	APIRootFn  func(path string) (cabby.APIRoot, error)
	APIRootsFn func() ([]cabby.APIRoot, error)
}

func (s APIRootService) APIRoot(path string) (cabby.APIRoot, error) {
	return s.APIRootFn(testAPIRootPath)
}

func (s APIRootService) APIRoots() ([]cabby.APIRoot, error) {
	return s.APIRootsFn()
}

type DiscoveryService struct {
	DiscoveryFn func() (cabby.Discovery, error)
}

func (s DiscoveryService) Discovery() (cabby.Discovery, error) {
	return s.DiscoveryFn()
}

type UserService struct {
	UserFn   func(user, password string) (cabby.User, error)
	ExistsFn func(cabby.User) bool
}

func (s UserService) User(user, password string) (cabby.User, error) {
	return s.UserFn(user, password)
}

func (s UserService) Exists(u cabby.User) bool {
	return s.ExistsFn(u)
}

/* helper functions */

func attemptRequest(c *http.Client, r *http.Request) (*http.Response, error) {
	fmt.Println("Requesting", r.URL, "from test server")
	MaxTries := 3

	for i := 0; i < MaxTries; i++ {
		res, err := c.Do(r)
		if err != nil || res != nil {
			return res, err
		}

		fmt.Println("Web server for test not responding, waiting...")
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

func mockHandler(testName string) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, testName)
	}
}

func mockHandlerFunc(testName string) http.Handler {
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

func testAPIRoot() cabby.APIRoot {
	return cabby.APIRoot{Path: testAPIRootPath,
		Title:            "test api root title",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
}

func testDiscovery() cabby.Discovery {
	return cabby.Discovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/",
		APIRoots:    []string{"test_api_root"}}
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

// create a context for the testUser and give it read/write access to the test collection
func withAuthentication(r *http.Request) *http.Request {
	ctx := context.WithValue(context.Background(), userName, testUserEmail)
	ctx = context.WithValue(ctx, canAdmin, true)
	return r.WithContext(ctx)
}
