package sqlite

import (
	"context"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestVersionServiceVersions(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.VersionService()

	result, err := s.Versions(context.Background(), tester.CollectionID, tester.ObjectID, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	expected := "2016-04-06T20:07:09Z"
	if result.Versions[0] != expected {
		t.Error("Got:", result.Versions[0], "Expected:", expected)
	}
}

func TestVersionServiceVersionQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.VersionService()

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Versions(context.Background(), tester.CollectionID, tester.ObjectID, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestVersionServiceVersionNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.VersionService()

	_, err := s.Versions(context.Background(), tester.CollectionID, tester.ObjectID, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
