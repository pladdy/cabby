package main

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	s "github.com/pladdy/stones"
)

/* helpers */

func malwareBundle() s.Bundle {
	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	bundleContent, _ := ioutil.ReadAll(bundleFile)

	var bundle s.Bundle
	json.Unmarshal(bundleContent, &bundle)

	return bundle
}

func writeMalwareBundle() {
	setupSQLite()

	bundle := malwareBundle()
	valid, errs := bundle.Validate()
	if !valid {
		fail.Fatal(errs)
	}

	// write the bundle
	ts := getStorer()
	defer ts.disconnect()
	writeBundle(bundle, testID, ts)
}

/* tests */

func TestStixObjectReadFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	so := stixObject{}
	err = so.read(ts, testID, "stixID")

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestStixObjectsRead(t *testing.T) {
	writeMalwareBundle()

	ts := getStorer()
	defer ts.disconnect()

	sos := stixObjects{}
	sos.read(ts, testID)

	if len(sos.Objects) != 3 {
		t.Error("Expected 3 objects")
	}
}

func TestStixObjectsReadFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	sos := stixObjects{}
	err = sos.read(ts, testID)

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestWriteBundle(t *testing.T) {
	setupSQLite()
	writeMalwareBundle()

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

func TestWriteBundleBadObject(t *testing.T) {
	setupSQLite()
	bundle := s.Bundle{}
	bundle.AddObject("invalid stix")

	// write the bundle
	ts := getStorer()
	defer ts.disconnect()
	writeBundle(bundle, testID, ts)

	// check for persistence
	s := getSQLiteDB()
	defer s.disconnect()

	expectedCount := 0
	var count int
	err := s.db.QueryRow("select count(*) from stix_objects where collection_id = '" + testID + "'").Scan(&count)
	if err != nil {
		t.Error(err)
	}

	if count != expectedCount {
		t.Error("Got:", count, "Expected:", expectedCount)
	}
}

func TestWriteBundleNoDuplicates(t *testing.T) {
	setupSQLite()

	// try to write the same bundle twice
	writeMalwareBundle()
	writeMalwareBundle()

	// check for persistence AND that only 3 objects are written (not 6 due to constraint on stix version)
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
