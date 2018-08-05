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
	testPort         = 1234
	testDiscoveryURL = "https://localhost:1234/taxii/"
	testUserEmail    = "test@cabby.com"
	testUserPassword = "test"
)

/* mock services */

type DiscoveryService struct {
	DiscoveryFn func() (cabby.Discovery, error)
}

func (s *DiscoveryService) Discovery() (cabby.Discovery, error) {
	return s.DiscoveryFn()
}

type UserService struct {
	UserFn  func(user, password string) (cabby.User, error)
	ValidFn func(cabby.User) bool
}

func (s *UserService) User(user, password string) (cabby.User, error) {
	return s.UserFn(user, password)
}

func (s *UserService) Valid(u cabby.User) bool {
	return s.ValidFn(u)
}

/* helper functions */

// handle generic testing of handlers.  It takes a handler function to call with a url;
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
