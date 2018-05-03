package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestHandleTaxiiManifest(t *testing.T) {
	setupSQLite()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	postBundle(u, "testdata/malware_bundle.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

	u = "https://localhost/api_root/collections/" + testID + "/manifest/"

	status, body := handlerTest(handleTaxiiManifest(ts), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var manifest taxiiManifest
	err := json.Unmarshal([]byte(body), &manifest)
	if err != nil {
		t.Fatal(err)
	}

	expected := 3
	if len(manifest.Objects) != expected {
		t.Error("Got:", len(manifest.Objects), "Expected:", expected)
	}
}

func TestHandleTaxiiManifestFail(t *testing.T) {
	defer setupSQLite()

	// drop the table so it can't be read
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/manifest/"

	status, _ := handlerTest(handleTaxiiManifest(ts), "GET", u, nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected", http.StatusNotFound)
	}
}

func TestHandleTaxiiManifestInvalidRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/manifest/"

	// create request and add a range to it that's invalid
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", u, nil))
	req.Header.Set("Range", "invalid range")

	res := httptest.NewRecorder()
	handleTaxiiManifest(ts)(res, req)

	if res.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Error("Got:", res.Code, "Expected:", http.StatusRequestedRangeNotSatisfiable)
	}
}

func TestHandleTaxiiManifestRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/manifest/"

	// create request and add a range to it that's invalid
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", u, nil))
	req.Header.Set("Range", "items 0-0")

	res := httptest.NewRecorder()
	handleTaxiiManifest(ts)(res, req)

	if res.Code != http.StatusPartialContent {
		t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
	}
}
