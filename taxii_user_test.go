package main

import (
	"testing"
)

func TestAssignedCollectionsFail(t *testing.T) {
	setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec(`drop table taxii_user_collection`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	ts := getStorer()
	defer ts.disconnect()

	_, err = assignedCollections(ts, testUser)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestNewTaxiiUser(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	// create a collection record
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`insert into taxii_collection (id, api_root_path, title, description, media_types)
	                    values ('` + id.String() + `', "api_root", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// associate user to collection
	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ('` + testUser + `', '` + id.String() + `', 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	user, err := newTaxiiUser(ts, testUser, testPass)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != testUser {
		t.Error("Got:", user.Email, "Expected:", testUser)
	}

	for _, v := range user.CollectionAccessList {
		if v.CanRead != true {
			t.Error("Got:", v.CanRead, "Expected:", true)
		}
		if v.CanWrite != true {
			t.Error("Got:", v.CanWrite, "Expected:", true)
		}
	}
}

func TestNewTaxiiUserFail(t *testing.T) {
	setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec(`drop table taxii_user`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	ts := getStorer()
	defer ts.disconnect()

	_, err = newTaxiiUser(ts, testUser, testPass)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiUserAssignedCollectionsReturnFail(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = newTaxiiUser(ts, testUser, testPass)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiUserCreateFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	tu := taxiiUser{}
	err = tu.create(ts)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestTaxiiUserReadFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	tu := taxiiUser{}
	err = tu.read(ts, "fail")
	if err == nil {
		t.Error("Expected error")
	}
}
