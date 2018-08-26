package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestDiscoveryServiceDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	expected := tester.DiscoveryDataStore

	result, err := s.Discovery()
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestDiscoveryServiceDiscoveryQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	_, err := ds.DB.Exec("drop table taxii_discovery")
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
