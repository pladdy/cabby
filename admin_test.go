// a lot of these tests are duplicative, but i'm doing that in the hope that it's more clear
// what's being tested.  that might be a dumb idea and instead there should be DRY'er tests below
package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"regexp"
	"strings"
	"testing"

	log "github.com/sirupsen/logrus"
)

var methodsToTest = []string{"DELETE", "POST", "PUT"}

var contextTests = []struct {
	contexts map[key]string
	status   int
}{
	{map[key]string{userName: testUser}, http.StatusForbidden},
	{map[key]string{}, http.StatusUnauthorized},
}

/* api root tests */

func TestHandleAdminTaxiiAPIRootInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()

	req := httptest.NewRequest("CUSTOM", testAdminAPIRootURL, nil)
	res := httptest.NewRecorder()
	handleAdminTaxiiAPIRoot(ts, handler)(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Error("Got:", res.Code, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiAPIRoot(t *testing.T) {
	setupSQLite()

	tests := []struct {
		method  string
		payload string
		title   string
	}{
		{method: "DELETE", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`, title: ""},
		{method: "POST", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`, title: t.Name()},
		{method: "PUT", payload: `{"path": "` + t.Name() + `", "title": "` + "updated" + `"}`, title: "updated"},
	}

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	for _, test := range tests {
		handler := http.NewServeMux()
		b := bytes.NewBuffer([]byte(test.payload))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), test.method, testAdminAPIRootURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK)
		}

		if test.title != "" {
			var title string
			err := s.db.QueryRow("select title from taxii_api_root where api_root_path = '" + t.Name() + "'").Scan(&title)
			if err != nil {
				t.Fatal(err)
			}

			if title != test.title {
				t.Error("Got:", title, "Expected:", test.title, "Method:", test.method)
			}
		}
	}
}

func TestAttemptRegisterAPIRoot(t *testing.T) {
	// post, delete, post and expect logging due to shared handler getting routes it already has
	setupSQLite()
	handler := http.NewServeMux()

	var buf bytes.Buffer

	log.SetOutput(&buf)
	defer func() {
		log.SetOutput(os.Stderr)
	}()

	tests := []struct {
		method  string
		payload string
	}{
		{method: "POST", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`},
		{method: "DELETE", payload: `{"path": "` + t.Name() + `", "title": "` + "updated" + `"}`},
		{method: "POST", payload: `{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.payload))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), test.method, testAdminAPIRootURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}
	}

	// check for warning in logs
	logs := regexp.MustCompile("\n").Split(strings.TrimSpace(buf.String()), -1)
	lastLog := logs[len(logs)-1]

	if match, _ := regexp.Match("failed to register api root handlers", []byte(lastLog)); !match {
		t.Error("Expected log output")
	}
}

func TestHandleAdminTaxiiAPIRootFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()
	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminAPIRootURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiAPIRoot(ts, handler)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiAPIRootFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		handler := http.NewServeMux()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), method, testAdminAPIRootURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiAPIRootFailServer(t *testing.T) {
	for _, method := range methodsToTest {
		setupSQLite()

		s := getSQLiteDB()
		defer s.disconnect()

		_, err := s.db.Exec("drop table taxii_api_root")
		if err != nil {
			t.Fatal(err)
		}

		ts := getStorer()
		defer ts.disconnect()

		handler := http.NewServeMux()

		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiAPIRoot(ts, handler), method, testAdminAPIRootURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* collections */

func TestHandleAdminTaxiiCollectionsInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	req := httptest.NewRequest("CUSTOM", testAdminCollectionsURL, nil)
	res := httptest.NewRecorder()
	handleAdminTaxiiCollections(ts)(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Error("Got:", res.Code, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiCollections(t *testing.T) {
	setupSQLite()
	basePayload := `{"id": "cd9552e7-fb5d-4628-a724-0772ed51200c", "api_root_path": "`

	tests := []struct {
		method  string
		payload string
		title   string
	}{
		{method: "DELETE", payload: basePayload + t.Name() + `", "title": "` + t.Name() + `"}`, title: ""},
		{method: "POST", payload: basePayload + t.Name() + `", "title": "` + t.Name() + `"}`, title: t.Name()},
		{method: "PUT", payload: basePayload + t.Name() + `", "title": "` + "updated" + `"}`, title: "updated"},
	}

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.payload))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), test.method, testAdminCollectionsURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}

		if test.title != "" {
			var title string
			err := s.db.QueryRow("select title from taxii_collection where api_root_path = '" + t.Name() + "'").Scan(&title)
			if err != nil {
				t.Fatal(err)
			}

			if title != test.title {
				t.Error("Got:", title, "Expected:", test.title, "Method:", test.method)
			}
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminCollectionsURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiCollections(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailData(t *testing.T) {
	setupSQLite()

	for _, method := range methodsToTest {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), method, testAdminCollectionsURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiCollectionsFailServer(t *testing.T) {
	for _, method := range methodsToTest {
		setupSQLite()

		s := getSQLiteDB()
		defer s.disconnect()

		_, err := s.db.Exec("drop table taxii_collection")
		if err != nil {
			t.Fatal(err)
		}

		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiCollections(ts), method, testAdminCollectionsURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}

/* discovery */

func TestHandleAdminTaxiiDiscoveryInvalidMethod(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	req := httptest.NewRequest("CUSTOM", testAdminDiscoveryURL, nil)
	res := httptest.NewRecorder()
	handleAdminTaxiiDiscovery(ts)(res, req)

	if res.Code != http.StatusMethodNotAllowed {
		t.Error("Got:", res.Code, "Expected:", http.StatusMethodNotAllowed)
	}
}

func TestHandleAdminTaxiiDiscovery(t *testing.T) {
	setupSQLite()

	tests := []struct {
		method  string
		payload string
		title   string
	}{
		{method: "DELETE", payload: `{"title": "` + t.Name() + `"}`, title: ""},
		{method: "POST", payload: `{"title": "` + t.Name() + `"}`, title: t.Name()},
		{method: "PUT", payload: `{"title": "` + "updated" + `"}`, title: "updated"},
	}

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	for _, test := range tests {
		b := bytes.NewBuffer([]byte(test.payload))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), test.method, testAdminDiscoveryURL, b)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected:", http.StatusOK, "Method:", test.method)
		}

		if test.title != "" {
			var title string
			err := s.db.QueryRow("select title from taxii_discovery").Scan(&title)
			if err != nil {
				t.Fatal(err)
			}

			if title != test.title {
				t.Error("Got:", title, "Expected:", test.title, "Method:", test.method)
			}
		}
	}
}

func TestHandleAdminTaxiiDiscoveryFailAuth(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range contextTests {
			ctx := testingContext()

			for k, v := range test.contexts {
				ctx = context.WithValue(ctx, k, v)
			}

			req := httptest.NewRequest(method, testAdminDiscoveryURL, b)
			req = req.WithContext(ctx)
			res := httptest.NewRecorder()
			handleAdminTaxiiDiscovery(ts)(res, req)

			if res.Code != test.status {
				t.Error("Got:", res.Code, "Expected:", test.status)
			}
		}
	}
}

func TestHandleAdminTaxiiDiscoveryFailData(t *testing.T) {
	setupSQLite()
	methods := []string{"PUT", "POST"}

	for _, method := range methods {
		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"pa"` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), method, testAdminDiscoveryURL, b)

		if status != http.StatusBadRequest {
			t.Error("Got:", status, "Expected:", http.StatusBadRequest)
		}
	}
}

func TestHandleAdminTaxiiDiscoveryDeleteFail(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec(`drop table taxii_discovery`)
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), "DELETE", testAdminDiscoveryURL, nil)

	if status != http.StatusInternalServerError {
		t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
	}
}

func TestHandleAdminTaxiiDiscoveryFailServer(t *testing.T) {
	for _, method := range methodsToTest {
		setupSQLite()

		s := getSQLiteDB()
		defer s.disconnect()

		_, err := s.db.Exec("drop table taxii_discovery")
		if err != nil {
			t.Fatal(err)
		}

		ts := getStorer()
		defer ts.disconnect()

		b := bytes.NewBuffer([]byte(`{"path": "` + t.Name() + `", "title": "` + t.Name() + `"}`))
		status, _ := handlerTest(handleAdminTaxiiDiscovery(ts), method, testAdminDiscoveryURL, b)

		if status != http.StatusInternalServerError {
			t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
		}
	}
}
