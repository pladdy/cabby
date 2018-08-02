package sqlite

import (
	"testing"

	cabby "github.com/pladdy/cabby2"
)

func TestDiscoveryServiceRead(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	expected := testDiscovery

	result, err := s.Read()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	discovery, ok := result.Data.(cabby.Discovery)
	if !ok {
		t.Fatal("Got:", result, "Expected Discovery")
	}

	if discovery.Title != expected.Title {
		t.Error("Got:", discovery.Title, "Expected:", expected.Title)
	}
	if discovery.Description != expected.Description {
		t.Error("Got:", discovery.Description, "Expected:", expected.Description)
	}
	if discovery.Contact != expected.Contact {
		t.Error("Got:", discovery.Contact, "Expected:", expected.Contact)
	}
	if discovery.Default != expected.Default {
		t.Error("Got:", discovery.Default, "Expected:", expected.Default)
	}
}

func TestDiscoveryServiceReadQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Read()
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestDiscoveryServiceReadNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	_, err := s.Read()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
