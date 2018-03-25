package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	"time"

	s "github.com/pladdy/stones"
)

func TestWriteBundle(t *testing.T) {
	setupSQLite()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundleContent, _ := ioutil.ReadAll(bundleFile)

	var bundle s.Bundle
	json.Unmarshal(bundleContent, &bundle)

	valid, errs := bundle.Validate()
	if !valid {
		t.Fatal(errs)
	}

	// write the bundle
	ts := getStorer()
	defer ts.disconnect()

	errsChannel := writeBundle(bundle, testID, ts)
	for err := range errsChannel {
		if err != nil {
			t.Error(err)
		}
	}

	time.Sleep(100 * time.Millisecond)

	// check for persistence
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

func TestNewStixObjectError(t *testing.T) {
	b, err := newStixObject([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for bundle:", b)
	}
}
