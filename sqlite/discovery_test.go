package sqlite

import (
	"context"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestUserServiceCreateDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	// there can only be one discovery
	s.DeleteDiscovery(context.Background())

	expected := tester.Discovery
	err := s.CreateDiscovery(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.Discovery(context.Background())
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceCreateDiscoveryInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	err := s.CreateDiscovery(context.Background(), cabby.Discovery{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceCreateDiscoveryQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	_, err := ds.DB.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateDiscovery(context.Background(), tester.Discovery)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceDeleteDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	// create and verify a user
	expected := tester.Discovery

	result, err := s.Discovery(context.Background())
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// delete and verify user is gone
	err = s.DeleteDiscovery(context.Background())
	if err != nil {
		t.Error("Got:", err)
	}

	result, err = s.Discovery(context.Background())
	if err != nil {
		t.Error("Got:", err)
	}

	if result.Title != "" {
		t.Error("Got:", result, `Expected: ""`)
	}
}

func TestUserServiceDeleteDiscoveryQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	_, err := ds.DB.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteDiscovery(context.Background())
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestDiscoveryServiceDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	expected := tester.DiscoveryDataStore

	result, err := s.Discovery(context.Background())
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

	_, err = s.Discovery(context.Background())
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestDiscoveryServiceDiscoveryNoDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := DiscoveryService{DB: ds.DB}

	_, err := s.Discovery(context.Background())
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}

func TestUserServiceUpdateDiscovery(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	expected := tester.Discovery
	expected.Description = "an updated description"

	err := s.UpdateDiscovery(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	// check it
	result, err := s.Discovery(context.Background())
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareDiscovery(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUserServiceUpdateDiscoveryInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	err := s.UpdateDiscovery(context.Background(), cabby.Discovery{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestUserServiceUpdateDiscoveryQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.DiscoveryService()

	_, err := ds.DB.Exec("drop table taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpdateDiscovery(context.Background(), tester.Discovery)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
