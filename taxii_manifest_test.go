package main

import (
	"encoding/json"
	"net/http"
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
