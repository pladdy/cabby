package main

import (
	"testing"
)

func TestRoutableCollectionsRead(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	rs := routableCollections{}

	err := rs.read(ts, testAPIRootPath)
	if err != nil {
		t.Error(err)
	}

	if len(rs.CollectionIDs) <= 0 {
		t.Error("Got 0 collections", "Expected at least one")
	}
}

func TestRoutableCollectionsReadFail(t *testing.T) {
	setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	rs := routableCollections{}

	err = rs.read(ts, testAPIRootPath)
	if err == nil {
		t.Error("Expected error")
	}
}
