package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestAPIRootServiceAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	result, err := s.APIRoot(expected.Path)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareAPIRoot(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestAPIRootServiceAPIRootQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	_, err := ds.DB.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoot(tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestAPIRootsServiceAPIRoots(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	result, err := s.APIRoots()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(result) <= 0 {
		t.Error("Got:", len(result), "Expected: > 0 results")
	}

	passed := tester.CompareAPIRoot(result[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestAPIRootsServiceAPIRootsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	_, err := ds.DB.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoots()
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
