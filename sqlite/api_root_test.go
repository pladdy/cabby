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
