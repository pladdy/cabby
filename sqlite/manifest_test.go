package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestManifestServiceManifest(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ManifestService{DB: ds.DB}

	expected := tester.ManifestEntry

	result, err := s.Manifest(tester.CollectionID)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareManifestEntry(result.Objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestManifestServiceManifestQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ManifestService{DB: ds.DB}

	_, err := s.DB.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Manifest(tester.CollectionID)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestManifestServiceManifestNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ManifestService{DB: ds.DB}

	_, err := s.Manifest(tester.CollectionID)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
