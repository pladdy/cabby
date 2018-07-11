package main

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestHandleTaxiiStatus(t *testing.T) {
	setupSQLite()

	returnedStatusID, err := postBundle(objectsURL(), "testdata/malware_bundle.json")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	// check the status
	u := testStatusURL + returnedStatusID

	status, body := handlerTest(handleTaxiiStatus(ts), "GET", u, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	var checkedStatus taxiiStatus
	err = json.Unmarshal([]byte(body), &checkedStatus)
	if err != nil {
		t.Fatal("Got:", err, "Expected a status resource")
	}
}

func TestHandleTaxiiStatusInvalidID(t *testing.T) {
	setupSQLite()

	_, err := postBundle(objectsURL(), "testdata/malware_bundle.json")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiStatus(ts), "GET", testStatusURL+"foo", nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
}

func TestHandleTaxiiStatusReadFail(t *testing.T) {
	setupSQLite()
	defer setupSQLite()

	returnedStatusID, err := postBundle(objectsURL(), "testdata/malware_bundle.json")
	if err != nil {
		t.Fatal(err)
	}

	// drop the table so it can't be read
	s := getSQLiteDB()
	defer s.disconnect()

	s.db.Exec("drop table taxii_status")

	// check the status
	ts := getStorer()
	defer ts.disconnect()

	u := testStatusURL + returnedStatusID

	status, _ := handlerTest(handleTaxiiStatus(ts), "GET", u, nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
}

func TestNewTaxiiStatusFail(t *testing.T) {
	_, err := newTaxiiStatus(0)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiStatusCompleted(t *testing.T) {
	tests := []struct {
		status   taxiiStatus
		expected bool
	}{
		{taxiiStatus{Status: "complete"}, true},
		{taxiiStatus{Status: "pending"}, false},
	}

	for _, test := range tests {
		if test.status.completed() != test.expected {
			t.Error("Got:", test.status.completed(), "Expected:", test.expected)
		}
	}
}
