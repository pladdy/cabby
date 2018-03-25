package main

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

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