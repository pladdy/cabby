package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestHandleTaxiiManifest(t *testing.T) {
	setupSQLite()

	u := "https://localhost/api_root/collections/" + testCollectionID + "/objects/"
	postBundle(u, "testdata/malware_bundle.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

	u = "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"

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

func TestHandleTaxiiManifestAddedAfter(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"
	tm := slowlyPosLasttBundle()
	u = u + "?added_after=" + tm.Format(time.RFC3339Nano)

	status, body := handlerTest(handleTaxiiManifest(ts), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var manifest taxiiManifest
	err := json.Unmarshal([]byte(body), &manifest)
	if err != nil {
		t.Fatal(err)
	}

	expected := 1
	if len(manifest.Objects) != expected {
		t.Error("Got:", len(manifest.Objects), "Expected:", expected)
	}
}

func TestHandleTaxiiManifestFilter(t *testing.T) {
	tests := []struct {
		filter  string
		objects int
	}{
		{"match[type]=indicator", 1},
		{"match[type]=indicator,malware", 2},
		{"match[type]=foo", 0},
		{"match[id]=indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f", 1},
		{"match[version]=2016-04-06T20:06:37.000Z", 1},
		{"match[version]=all", 4},
		{"match[version]=first", 3},
		{"match[version]=foo", 3}, // invalid, defaults to 'last'
		// composite filters
		{"match[type]=indicator,malware&match[version]=all", 3},
	}

	setupSQLite()
	postBundle(objectsURL(), "testdata/multi_filter.json")

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		u := "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"
		u = u + "?" + test.filter

		status, body := attemptHandlerTest(handleTaxiiManifest(ts), "GET", u, nil)

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected", http.StatusOK, "Filter:", test.filter)
		}

		var manifest taxiiManifest
		err := json.Unmarshal([]byte(body), &manifest)
		if err != nil {
			t.Fatal(err)
		}

		if len(manifest.Objects) != test.objects {
			t.Error("Got:", len(manifest.Objects), "Expected:", test.objects, "Filter:", test.filter)
		}
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

	u := "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"

	status, _ := handlerTest(handleTaxiiManifest(ts), "GET", u, nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected", http.StatusNotFound)
	}
}

func TestHandleTaxiiManifestInvalidRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"

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

	u := "https://localhost/api_root/collections/" + testCollectionID + "/manifest/"

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
