package sqlite

import (
	"strings"
	"testing"
)

func TestAPIRootServiceRead(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := APIRootService{DB: ds.DB}

	expected := testAPIRoot

	result, err := s.APIRoot(testAPIRootPath)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Path != expected.Path {
		t.Error("Got:", result.Path, "Expected:", expected.Path)
	}
	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		t.Error("Got:", result.Description, "Expected:", expected.Description)
	}
	if strings.Join(result.Versions, ",") != strings.Join(expected.Versions, ",") {
		t.Error("Got:", strings.Join(result.Versions, ","), "Expected:", strings.Join(expected.Versions, ","))
	}
	if result.MaxContentLength != expected.MaxContentLength {
		t.Error("Got:", result.MaxContentLength, "Expected:", expected.MaxContentLength)
	}
}

func TestAPIRootServiceReadQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := APIRootService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoot(testAPIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestAPIRootsServiceRead(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := APIRootService{DB: ds.DB}

	expected := testAPIRoot

	result, err := s.APIRoots()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(result) <= 0 {
		t.Error("Got:", len(result), "Expected: > 0 results")
	}

	if result[0].Path != expected.Path {
		t.Error("Got:", result[0].Path, "Expected:", expected.Path)
	}
	if result[0].Title != expected.Title {
		t.Error("Got:", result[0].Title, "Expected:", expected.Title)
	}
	if result[0].Description != expected.Description {
		t.Error("Got:", result[0].Description, "Expected:", expected.Description)
	}
	if strings.Join(result[0].Versions, ",") != strings.Join(expected.Versions, ",") {
		t.Error("Got:", strings.Join(result[0].Versions, ","), "Expected:", strings.Join(expected.Versions, ","))
	}
	if result[0].MaxContentLength != expected.MaxContentLength {
		t.Error("Got:", result[0].MaxContentLength, "Expected:", expected.MaxContentLength)
	}
}

func TestAPIRootsServiceReadQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := APIRootService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.APIRoots()
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
