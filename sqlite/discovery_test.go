package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestDiscoveryServiceDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	expected := tester.Discovery

	result, err := s.Discovery()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		t.Error("Got:", result.Description, "Expected:", expected.Description)
	}
	if result.Contact != expected.Contact {
		t.Error("Got:", result.Contact, "Expected:", expected.Contact)
	}
	if result.Default != expected.Default {
		t.Error("Got:", result.Default, "Expected:", expected.Default)
	}
}

func TestDiscoveryServiceDiscoveryQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Discovery()
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestDiscoveryServiceDiscoveryNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	_, err := s.Discovery()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
