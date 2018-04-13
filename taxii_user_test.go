package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

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

	for _, v := range user.CollectionAccess {
		if v.CanRead != true {
			t.Error("Got:", v.CanRead, "Expected:", true)
		}
		if v.CanWrite != true {
			t.Error("Got:", v.CanWrite, "Expected:", true)
		}
	}
}

func TestNewTaxiiUserNoAccess(t *testing.T) {
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

	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(testPass)))
	_, err = newTaxiiUser(ts, testUser, pass)
	if err == nil {
		t.Error("Expected error with no access")
	}
}

func TestNewTaxiiUserFail(t *testing.T) {
	defer loadTestConfig()

	ts := getStorer()
	defer ts.disconnect()

	config = Config{}

	_, err := newTaxiiUser(ts, "test@test.fail", "nopass")
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

func TestVerifyValidUserFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	b, err := verifyValidUser(ts, "fail", "test")

	if b != false {
		t.Error("Got:", b, "Expected: false")
	}

	if err == nil {
		t.Error("Expected error")
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
	err = tu.create(ts, "fail")
	if err == nil {
		t.Error("Expected error")
	}
}
