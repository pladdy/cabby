package main

import (
	"encoding/json"
	"testing"
)

func TestHandleTaxiiAPIRoot(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, result := handlerTest(handleTaxiiAPIRoot(ts), "GET", testAPIRootURL, nil)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
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

func TestHandleTaxiiAPIRootFailRead(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiAPIRoot.sql", "backend/sqlite/read/taxiiAPIRoot.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiAPIRoot.sql.testing", "backend/sqlite/read/taxiiAPIRoot.sql")

	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiAPIRoot(ts), "GET", testAPIRootURL, nil)

	if status != 400 {
		t.Error("Got:", status, "Expected: 400")
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

	if status != 404 {
		t.Error("Got:", status, "Expected: 400")
	}
}
