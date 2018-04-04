package main

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	s "github.com/pladdy/stones"
)

/* helpers */
func postBundle(u, bundlePath string) {
	ts := getStorer()
	defer ts.disconnect()

	// post a bundle to the data store
	bundleFile, _ := os.Open(bundlePath)
	bundleContent, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundleContent)
	handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	// give time for bundle to be persisted
	time.Sleep(100 * time.Millisecond)
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

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
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

func TestHandleTaxiiObjectGetMultipleVersions(t *testing.T) {
	setupSQLite()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	postBundle(u, "testdata/multiple_versions.json")

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

	if len(bundle.Objects) != 2 {
		t.Error("Got:", len(bundle.Objects), "Expected 2 objects")
	}
}

func TestHandleTaxiiObjectsGet(t *testing.T) {
	setupSQLite()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	postBundle(u, "testdata/malware_bundle.json")

	// read the bundle back
	ts := getStorer()
	defer ts.disconnect()

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

	if len(bundle.Objects) != 3 {
		t.Error("Expected 3 objects")
	}
}

func TestHandleTaxiiObjectsGetCollectionUnauthorized(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	maxContent := int64(2048)

	// omit a context with the testUser name in it
	req := httptest.NewRequest("GET", u, nil)
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

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected", http.StatusOK)
	}

	// posting bundles is asynchronous, and when you post a status resource is returned
	// writes are done in the backaground at that point and we need a wait to let them persist
	// note to future self: this implies a performance requirement of 3 writes within 100 ms?
	time.Sleep(100 * time.Millisecond)

	s := getSQLiteDB()
	defer s.disconnect()

	expectedCount := 3
	var count int
	err := s.db.QueryRow("select count(*) from stix_objects where collection_id = '" + testID + "'").Scan(&count)
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

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(1)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if status != http.StatusRequestEntityTooLarge {
		t.Error("Got:", status, "Expected:", http.StatusRequestEntityTooLarge)
	}
}

func TestHandleTaxiiObjectsPostInvalidBundle(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundle := []byte(`{"foo":"bar"}`)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestHandleTaxiiObjectsMethodNotAllowed(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	maxContent := int64(1)
	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	status, _ := handlerTest(handleTaxiiObjects(ts, maxContent), "PUT", u, nil)

	if status != http.StatusMethodNotAllowed {
		t.Error("Expected status to be:", http.StatusMethodNotAllowed)
	}
}

func TestHandleTaxiiObjectPostCollectionUnauthorized(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u := "https://localhost/api_root/collections/" + testID + "/objects/"
	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundle, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundle)

	// omit a context with the testUser name in it
	req := httptest.NewRequest("POST", u, b)
	res := httptest.NewRecorder()
	h := handleTaxiiObjects(ts, maxContent)
	h(res, req)

	if res.Code != http.StatusUnauthorized {
		t.Error("Got:", res.Code, "Expected:", http.StatusUnauthorized)
	}
}
