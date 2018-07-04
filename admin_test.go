package main

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

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

		var title string
		err := s.db.QueryRow("select title from taxii_api_root where api_root_path = '" + t.Name() + "'").Scan(&title)
		if err != nil {
			t.Fatal(err)
		}

		if title != test.title {
			t.Error("Got:", title, "Expected:", test.title)
		}
	}
}

func TestHandleAdminTaxiiAPIRootFailAuth(t *testing.T) {
	setupSQLite()

	methodsToTest := []string{"POST", "PUT"}

	tests := []struct {
		contexts map[key]string
		status   int
	}{
		{map[key]string{userName: testUser}, http.StatusForbidden},
		{map[key]string{}, http.StatusUnauthorized},
	}

	ts := getStorer()
	defer ts.disconnect()

	handler := http.NewServeMux()
	b := bytes.NewBuffer([]byte(`{"this won't get processed": true}`))

	for _, method := range methodsToTest {
		for _, test := range tests {
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

	methodsToTest := []string{"POST", "PUT"}

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
	methodsToTest := []string{"POST", "PUT"}

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
