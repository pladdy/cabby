package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestUserServiceUser(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	expected := tester.User

	result, err := s.User(tester.Context, tester.UserPassword)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Email != expected.Email {
		t.Error("Got:", result.Email, "Expected:", expected.Email)
	}
	if result.CanAdmin != expected.CanAdmin {
		t.Error("Got:", result.CanAdmin, "Expected:", expected.CanAdmin)
	}
}

func TestUserServiceUserQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.User(tester.Context, tester.UserPassword)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUserServiceUserCollections(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	expected := tester.UserCollectionList

	result, err := s.UserCollections(tester.Context)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.Email != expected.Email {
		t.Error("Got:", result.Email, "Expected:", expected.Email)
	}

	for id, ca := range result.CollectionAccessList {
		if ca.CanRead != expected.CollectionAccessList[id].CanRead {
			t.Error("Got:", ca.CanRead, "Expected:", expected.CollectionAccessList[id].CanRead)
		}
	}
}

func TestUserServiceUserCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.UserService()

	_, err := ds.DB.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.UserCollections(tester.Context)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
