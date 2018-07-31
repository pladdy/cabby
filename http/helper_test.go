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
	testDiscovery    = cabby.Discovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
)

// handle generic testing of handlers.  It takes a handler function to call with a url;
// it returns the status code and response as a string
func handlerTest(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	var req *http.Request

	if b != nil {
		req = withAuthContext(httptest.NewRequest(method, url, b))
	} else {
		req = withAuthContext(httptest.NewRequest(method, url, nil))
	}

	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

// create a context for the testUser and give it read/write access to the test collection
func withAuthContext(r *http.Request) *http.Request {
	ctx := context.WithValue(testingContext(), userName, testUser)
	ctx = context.WithValue(ctx, canAdmin, true)
	return r.WithContext(ctx)
}
