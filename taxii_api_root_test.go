package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHandleTaxiiAPIRoot(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, result := handlerTest(handleTaxiiAPIRoot(ts), "GET", testAPIRootURL, nil)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var ta taxiiAPIRoot
	err := json.Unmarshal([]byte(result), &ta)
	if err != nil {
		t.Fatal(err)
	}

	if ta.Title != testAPIRoot.Title {
		t.Error("Got:", ta.Title, "Expected:", testAPIRoot.Title)
	}
}

func TestHandleTaxiiAPIRootReadFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_api_root")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiAPIRoot(ts), "GET", testAPIRootURL, nil)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}
}

func TestHandleTaxiiAPIRootNotFound(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	s.db.Exec("delete from taxii_api_root")

	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiAPIRoot(ts), "GET", testAPIRootURL, nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
}
