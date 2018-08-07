package http

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"

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

/* mock services */

type APIRootService struct {
	APIRootFn func(path string) (cabby.APIRoot, error)
}

func (s *APIRootService) APIRoot(path string) (cabby.APIRoot, error) {
	return s.APIRootFn(testAPIRootPath)
}

type DiscoveryService struct {
	DiscoveryFn func() (cabby.Discovery, error)
}

func (s *DiscoveryService) Discovery() (cabby.Discovery, error) {
	return s.DiscoveryFn()
}

type UserService struct {
	UserFn   func(user, password string) (cabby.User, error)
	ExistsFn func(cabby.User) bool
}

func (s *UserService) User(user, password string) (cabby.User, error) {
	return s.UserFn(user, password)
}

func (s *UserService) Exists(u cabby.User) bool {
	return s.ExistsFn(u)
}

/* helper functions */

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

// create a context for the testUser and give it read/write access to the test collection
func withAuthentication(r *http.Request) *http.Request {
	ctx := context.WithValue(context.Background(), userName, testUserEmail)
	ctx = context.WithValue(ctx, canAdmin, true)
	return r.WithContext(ctx)
}
