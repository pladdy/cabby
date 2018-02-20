package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHandleTaxiiDiscovery(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, result := handlerTest(handleTaxiiDiscovery(ts), "GET", discoveryURL, nil)

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	var td taxiiDiscovery
	err := json.Unmarshal([]byte(result), &td)
	if err != nil {
		t.Fatal(err)
	}

	if td.Default != discoveryURL {
		t.Error("Got:", td.Default, "Expected:", discoveryURL)
	}
}

func TestHandleTaxiiDiscoveryNoDiscovery(t *testing.T) {
	defer setupSQLite()

	// delete discovery from table
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("delete from taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	// now try to use handler
	ts := getStorer()
	defer ts.disconnect()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	h := handleTaxiiDiscovery(ts)
	h(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleTaxiiDiscoveryError(t *testing.T) {
	defer setupSQLite()

	// drop the table all together
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	// now try to use handler
	ts := getStorer()
	defer ts.disconnect()

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	h := handleTaxiiDiscovery(ts)
	h(res, req)

	if res.Code != 400 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestTaxiiDiscoveryFailParse(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiDiscovery.sql", "backend/sqlite/read/taxiiDiscovery.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiDiscovery.sql.testing", "backend/sqlite/read/taxiiDiscovery.sql")

	ts := getStorer()
	defer ts.disconnect()

	td := taxiiDiscovery{}
	err := td.read(ts)

	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}
