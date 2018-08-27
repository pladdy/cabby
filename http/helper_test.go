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
	"github.com/pladdy/stones"
)

var (
	testAPIRootURL     = tester.BaseURL + tester.APIRootPath + "/"
	testCollectionsURL = testAPIRootURL + "collections/"
	testCollectionURL  = testCollectionsURL + tester.CollectionID + "/"
	testManifestURL    = testCollectionURL + "manifest/"
	testObjectsURL     = testCollectionURL + "objects/"
	testObjectURL      = testObjectsURL + tester.ObjectID + "/"
	testStatusURL      = testAPIRootURL + "status/" + tester.StatusID + "/"
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

func callHandler(h http.HandlerFunc, req *http.Request) (int, string) {
	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

func getResponse(req *http.Request, server *httptest.Server) (*http.Response, error) {
	c := server.Client()
	return c.Do(req)
}

// test a HandlerFunc.  given a HandlerFunc, method, url, and bytes.Buffer, call the request and record response.
// it returns the status code and response as a string
func handlerTest(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	return callHandler(h, withAuthentication(newRequest(method, url, b)))
}

func handlerTestNoAuth(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	return callHandler(h, newRequest(method, url, b))
}

func lastLog(buf bytes.Buffer) string {
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	return logs[len(logs)-1]
}

func newRequest(method, url string, b *bytes.Buffer) *http.Request {
	if b != nil {
		return httptest.NewRequest(method, url, b)
	}
	return httptest.NewRequest(method, url, nil)
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
	ctx = context.WithValue(ctx, collectionAccessList, tester.UserCollectionList.CollectionAccessList)
	return r.WithContext(ctx)
}

/* mock services */

// mock services by default return no error and an empty value
// tests using them can manipulate them further for different test cases

func mockAPIRootService() tester.APIRootService {
	as := tester.APIRootService{}
	as.APIRootFn = func(path string) (cabby.APIRoot, error) { return tester.APIRoot, nil }
	as.APIRootsFn = func() ([]cabby.APIRoot, error) { return []cabby.APIRoot{tester.APIRoot}, nil }
	return as
}

func mockCollectionService() tester.CollectionService {
	cs := tester.CollectionService{}
	cs.CollectionFn = func(user, collectionID, apiRootPath string) (cabby.Collection, error) {
		return tester.Collection, nil
	}
	cs.CollectionsFn = func(user, apiRootPath string) (cabby.Collections, error) { return tester.Collections, nil }
	cs.CollectionsInAPIRootFn = func(apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
		return tester.CollectionsInAPIRoot, nil
	}
	return cs
}

func mockDiscoveryService() tester.DiscoveryService {
	ds := tester.DiscoveryService{}
	ds.DiscoveryFn = func() (cabby.Discovery, error) { return tester.Discovery, nil }
	return ds
}

func mockManifestService() tester.ManifestService {
	ms := tester.ManifestService{}
	ms.ManifestFn = func(collectionID string) (cabby.Manifest, error) { return tester.Manifest, nil }
	return ms
}

func mockObjectService() tester.ObjectService {
	osv := tester.ObjectService{}
	osv.CreateBundleFn = func(b stones.Bundle, collectionID string, s cabby.Status, ss cabby.StatusService) {
		tester.Info.Println("mock Creating Bundle")
	}
	osv.ObjectFn = func(collectionID, stixID string) (cabby.Object, error) { return tester.Object, nil }
	osv.ObjectsFn = func(collectionID string) ([]cabby.Object, error) { return tester.Objects, nil }
	return osv
}

func mockStatusService() tester.StatusService {
	ss := tester.StatusService{}
	ss.CreateStatusFn = func(status cabby.Status) error { return nil }
	ss.StatusFn = func(statusID string) (cabby.Status, error) { return tester.Status, nil }
	ss.UpdateStatusFn = func(status cabby.Status) error { return nil }
	return ss
}

func mockUserService() tester.UserService {
	us := tester.UserService{}
	us.UserFn = func(user, password string) (cabby.User, error) {
		return cabby.User{}, nil
	}
	us.UserCollectionsFn = func(user string) (cabby.UserCollectionList, error) {
		return cabby.UserCollectionList{}, nil
	}
	us.ExistsFn = func(cabby.User) bool { return true }
	return us
}

func mockDataStore() tester.DataStore {
	md := tester.DataStore{}
	md.APIRootServiceFn = func() tester.APIRootService { return mockAPIRootService() }
	md.CollectionServiceFn = func() tester.CollectionService { return mockCollectionService() }
	md.DiscoveryServiceFn = func() tester.DiscoveryService { return mockDiscoveryService() }
	md.ManifestServiceFn = func() tester.ManifestService { return mockManifestService() }
	md.ObjectServiceFn = func() tester.ObjectService { return mockObjectService() }
	md.StatusServiceFn = func() tester.StatusService { return mockStatusService() }
	md.UserServiceFn = func() tester.UserService { return mockUserService() }

	return md
}
