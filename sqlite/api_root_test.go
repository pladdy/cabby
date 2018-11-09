package sqlite

import (
	"context"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestAPIRootServiceAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	result, err := s.APIRoot(context.Background(), expected.Path)
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

	_, err := ds.DB.Exec("drop table api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoot(context.Background(), tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestAPIRootsServiceAPIRoots(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	result, err := s.APIRoots(context.Background())
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

	_, err := ds.DB.Exec("drop table api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoots(context.Background())
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceCreateAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot
	expected.Path = "new path"

	err := s.CreateAPIRoot(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.APIRoot(context.Background(), expected.Path)
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareAPIRoot(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceCreateAPIRootInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	err := s.CreateAPIRoot(context.Background(), cabby.APIRoot{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceCreateAPIRootQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	_, err := ds.DB.Exec("drop table api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateAPIRoot(context.Background(), tester.APIRoot)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceDeleteAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	// delete and verify user is gone
	err := s.DeleteAPIRoot(context.Background(), expected.Path)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.APIRoot(context.Background(), expected.Path)
	if err != nil {
		t.Error("Got:", err)
	}

	if result.Path != "" {
		t.Error("Got:", result, `Expected: ""`)
	}
}

func TestUserServiceDeleteAPIRootQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	_, err := ds.DB.Exec("drop table api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteAPIRoot(context.Background(), "foo")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceUpdateAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	expected := tester.APIRoot

	// update description
	expected.Description = "a description"

	err := s.UpdateAPIRoot(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	// check it
	result, err := s.APIRoot(context.Background(), expected.Path)
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareAPIRoot(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceUpdateAPIRootInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	err := s.UpdateAPIRoot(context.Background(), cabby.APIRoot{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceUpdateAPIRootQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.APIRootService()

	_, err := ds.DB.Exec("drop table api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpdateAPIRoot(context.Background(), tester.APIRoot)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
