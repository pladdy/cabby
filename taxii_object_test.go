package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"testing"
	"time"

	s "github.com/pladdy/stones"
)

func objectsURL() string {
	u, err := url.Parse(testObjectsURL)
	if err != nil {
		fail.Fatal(err)
	}
	return u.String()
}

func TestBundleFromBytesUnmarshalFail(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

func TestBundleFromBytesInvalidBundle(t *testing.T) {
	b, err := bundleFromBytes([]byte(`{"foo": "bar"}`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}

func TestHandleTaxiiObjectGet(t *testing.T) {
	setupSQLite()

	u := objectsURL()
	postBundle(u, "testdata/malware_bundle.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

	stixID := "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"
	u = u + stixID
	maxContent := int64(2048)

	status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var bundle s.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) != 1 {
		t.Error("Expected 1 object")
	}
}

func TestHandleTaxiiObjectGetFilter(t *testing.T) {
	tests := []struct {
		filter      string
		objects     int
		shouldError bool
	}{
		{"match[version]=2016-06-06T20:03:48.000Z", 1, false},
		{"match[version]=all", 3, false},
		{"match[version]=first", 1, false},
		{"match[version]=foo", 1, false}, // invalid, defaults to 'last'
	}

	setupSQLite()
	postBundle(objectsURL(), "testdata/multiple_versions.json")

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		maxContent := int64(2048)

		u := objectsURL()
		stixID := "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"
		u = u + stixID
		u = u + "/" + "?" + test.filter

		status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)

		if test.shouldError {
			if status != http.StatusNotFound {
				t.Error("Got:", status, "Expected", http.StatusNotFound)
			}
			continue
		}

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected", http.StatusOK, "Filter:", test.filter)
		}

		var bundle s.Bundle
		err := json.Unmarshal([]byte(body), &bundle)
		if err != nil {
			t.Fatal(err)
		}

		if len(bundle.Objects) != test.objects {
			t.Error("Got:", len(bundle.Objects), "Expected:", test.objects)
		}
	}
}

func TestHandleTaxiiObjectsGetFailNoObjects(t *testing.T) {
	setupSQLite()

	// drop the table so it can't be read
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	// access the url using the handler
	ts := getStorer()
	defer ts.disconnect()

	u := objectsURL()
	stixID := "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"
	u = u + stixID

	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", u, nil))
	res := httptest.NewRecorder()

	handleGetTaxiiObjects(ts, res, req)

	if res.Code != http.StatusNotFound {
		t.Error("Got:", res.Code, "Expected", http.StatusNotFound)
	}
}

func TestHandleTaxiiObjectsGetFailNoBundle(t *testing.T) {
	setupSQLite()

	u := "https://localhost/api_root/collections/" + testCollectionID + "/objects/"
	ts := getStorer()
	defer ts.disconnect()

	stixID := "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"
	u = u + stixID
	maxContent := int64(2048)

	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected", http.StatusNotFound)
	}
}

func TestHandleTaxiiObjectGetMultipleVersions(t *testing.T) {
	setupSQLite()
	postBundle(objectsURL(), "testdata/multiple_versions.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

	u := objectsURL()
	stixID := "indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f"
	u = u + stixID
	maxContent := int64(4096)

	status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var bundle s.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) != 1 {
		t.Error("Got:", len(bundle.Objects), "Expected 1 object")
	}
}

func TestHandleTaxiiObjectsGetInvalidRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	// create request and add a range to it that's invalid
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", objectsURL(), nil))
	req.Header.Set("Range", "invalid range")

	res := httptest.NewRecorder()
	handleTaxiiObjects(ts, 2048)(res, req)

	if res.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Error("Got:", res.Code, "Expected:", http.StatusRequestedRangeNotSatisfiable)
	}
}

func TestHandleTaxiiObjectsGetRange(t *testing.T) {
	setupSQLite()
	postBundle(objectsURL(), "testdata/multiple_versions.json")

	ts := getStorer()
	defer ts.disconnect()

	// create request and add a range to it that's invalid
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", objectsURL(), nil))
	req.Header.Set("Range", "items 0-0")

	res := httptest.NewRecorder()
	handleTaxiiObjects(ts, 2048)(res, req)

	if res.Code != http.StatusPartialContent {
		t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
	}

	body, _ := ioutil.ReadAll(res.Body)

	var bundle s.Bundle
	err := json.Unmarshal(body, &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) != 1 {
		t.Error("Got:", len(bundle.Objects), "Expected: 1")
	}
}

func TestHandleTaxiiObjectsGet(t *testing.T) {
	setupSQLite()
	postBundle(objectsURL(), "testdata/malware_bundle.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

	maxContent := int64(2048)
	status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", objectsURL(), nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var bundle s.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) != 3 {
		t.Error("Expected 3 objects")
	}
}

func TestHandleTaxiiObjectsGetAddedAfter(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	maxContent := int64(2048)
	tm := slowlyPostLastBundle()
	u := objectsURL() + "?added_after=" + tm.Format(time.RFC3339Nano)

	status, body := attemptHandlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	var bundle s.Bundle
	err := json.Unmarshal([]byte(body), &bundle)
	if err != nil {
		t.Fatal(err)
	}

	if len(bundle.Objects) != 1 {
		t.Error("Got:", len(bundle.Objects), "Expected: 1")
	}
}

func TestHandleTaxiiObjectsGetFilter(t *testing.T) {
	tests := []struct {
		filter      string
		objects     int
		shouldError bool
	}{
		{"match[type]=indicator", 1, false},
		{"match[type]=indicator,malware", 2, false},
		{"match[type]=foo", 0, true},
		{"match[id]=indicator--8e2e2d2b-17d4-4cbf-938f-98ee46b3cd3f", 1, false},
		{"match[version]=2016-04-06T20:06:37.000Z", 1, false},
		{"match[version]=all", 4, false},
		{"match[version]=first", 3, false},
		{"match[version]=foo", 3, false}, // invalid, defaults to 'last'
		// composite filters
		{"match[type]=indicator,malware&match[version]=all", 3, false},
	}

	setupSQLite()
	postBundle(objectsURL(), "testdata/multi_filter.json")

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		maxContent := int64(2048)
		u := objectsURL() + "?" + test.filter

		status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "GET", u, nil)

		if test.shouldError {
			if status != http.StatusNotFound {
				t.Error("Got:", status, "Expected", http.StatusNotFound, "Filter:", test.filter)
			}
			continue
		}

		if status != http.StatusOK {
			t.Error("Got:", status, "Expected", http.StatusOK, "Filter:", test.filter)
		}

		var bundle s.Bundle
		err := json.Unmarshal([]byte(body), &bundle)
		if err != nil {
			t.Fatal(err)
		}

		if len(bundle.Objects) != test.objects {
			t.Error("Got:", len(bundle.Objects), "Expected:", test.objects, "Filter:", test.filter)
		}
	}
}

func TestHandleTaxiiObjectsGetCollectionUnauthorized(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	maxContent := int64(2048)

	// omit a context with the testUser name in it
	req := httptest.NewRequest("GET", objectsURL(), nil)
	res := httptest.NewRecorder()
	h := handleTaxiiObjects(ts, maxContent)
	h(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Error("Got:", res.Code, "Expected:", http.StatusUnauthorized)
	}
}

func TestHandleTaxiiObjectsPost(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", objectsURL(), b)

	if status != http.StatusAccepted {
		t.Error("Got:", status, "Expected", http.StatusAccepted)
	}

	// posting bundles is asynchronous, and when you post a status resource is returned
	// writes are done in the backaground at that point and we need a wait to let them persist
	// note to future self: this implies a performance requirement of 3 writes within 100 ms?
	time.Sleep(100 * time.Millisecond)

	s := getSQLiteDB()
	defer s.disconnect()

	expectedCount := 3
	var count int
	err := s.db.QueryRow("select count(*) from stix_objects where collection_id = '" + testCollectionID + "'").Scan(&count)
	if err != nil {
		t.Error(err)
	}

	if count != expectedCount {
		t.Error("Got:", count, "Expected:", expectedCount)
	}
}

func TestHandleTaxiiObjectsPostContentTooBig(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(1)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", objectsURL(), b)

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestHandleTaxiiObjectsPostInvalidBundle(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	bundle := []byte(`{"foo":"bar"}`)
	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", objectsURL(), b)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestHandleTaxiiObjectsPostEmptyBundle(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	bundle := []byte(`{"type": "bundle", "objects": [], "spec_version": "2.0", "id": "bundle--5d0092c5-5f74-4287-9642-33f4c354e56d"}`)
	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", objectsURL(), b)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestHandleTaxiiObjectsPostServerError(t *testing.T) {
	defer setupSQLite()

	db := getSQLiteDB()
	defer db.disconnect()

	db.db.Exec("drop table taxii_status")

	ts := getStorer()
	defer ts.disconnect()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", objectsURL(), b)

	if status != http.StatusInternalServerError {
		t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
	}
}

func TestHandleTaxiiObjectsMethodNotAllowed(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	maxContent := int64(1)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "PUT", objectsURL(), nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Expected status to be:", http.StatusMethodNotAllowed)
	}
}

func TestHandleTaxiiObjectPostCollectionUnauthorized(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)

	// omit a context with the testUser name in it
	req := httptest.NewRequest("POST", objectsURL(), b)
	res := httptest.NewRecorder()
	h := handleTaxiiObjects(ts, maxContent)
	h(res, req)

	if res.Code != http.StatusForbidden {
		t.Error("Got:", res.Code, "Expected:", http.StatusForbidden)
	}
}
