package main

import (
	"encoding/json"
	"net/http/httptest"
	"testing"
)

func TestHandleTaxiiDiscovery(t *testing.T) {
	status, result := handlerTest(handleTaxiiDiscovery, "GET", discoveryURL, nil)

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

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("delete from taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 404 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestHandleTaxiiDiscoveryError(t *testing.T) {
	defer setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", discoveryURL, nil)
	res := httptest.NewRecorder()
	handleTaxiiDiscovery(res, req)

	if res.Code != 400 {
		t.Error("Got:", res.Code, "Expected:", 404)
	}
}

func TestTaxiiDiscoveryFailTaxiiStorer(t *testing.T) {
	config = cabbyConfig{}
	defer loadTestConfig()

	td := taxiiDiscovery{}
	err := td.read()

	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiDiscoveryFailParse(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiDiscovery.sql", "backend/sqlite/read/taxiiDiscovery.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiDiscovery.sql.testing", "backend/sqlite/read/taxiiDiscovery.sql")

	td := taxiiDiscovery{}
	err := td.read()

	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiDiscoveryFailRead(t *testing.T) {
	defer setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	td := taxiiDiscovery{}
	err = td.read()

	if err == nil {
		t.Error("Expected error")
	}
}
