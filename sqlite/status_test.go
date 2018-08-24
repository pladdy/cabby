package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestStatusServiceCreateStatus(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := StatusService{DB: ds.DB, DataStore: ds}

	test := tester.Status

	err := s.CreateStatus(test)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.Status(test.ID.String())
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareStatus(result, test)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestStatusServiceStatus(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := StatusService{DB: ds.DB}

	expected := tester.Status

	result, err := s.Status(expected.ID.String())
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestStatusServiceStatusQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := StatusService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_status")
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Status

	_, err = s.Status(expected.ID.String())
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestStatusServiceUpdateStatus(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := StatusService{DB: ds.DB, DataStore: ds}

	// create a status
	test := tester.Status
	err := s.CreateStatus(test)
	if err != nil {
		t.Error("Got:", err)
	}

	// update the status
	test.SuccessCount = 1
	test.PendingCount = 3
	test.FailureCount = 1

	err = s.UpdateStatus(test)
	if err != nil {
		t.Error("Got:", err)
	}

	// verify it's updated
	result, err := s.Status(test.ID.String())
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareStatus(result, test)
	if !passed {
		t.Error("Comparison failed")
	}
}
