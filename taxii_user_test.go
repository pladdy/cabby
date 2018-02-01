package main

import (
	"crypto/sha256"
	"fmt"
	"testing"
)

func TestNewTaxiiUser(t *testing.T) {
	setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ('` + id.String() + `', "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// associate user to collection
	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ('` + testUser + `', '` + id.String() + `', 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	user, err := newTaxiiUser(testUser, testPass)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != testUser {
		t.Error("Got:", user.Email, "Expected:", testUser)
	}

	for k, v := range user.CollectionAccess {
		info.Println("Got collection id of", k)

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

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ('` + id.String() + `', "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(testPass)))
	_, err = newTaxiiUser(testUser, pass)
	if err == nil {
		t.Error("Expected error with no access")
	}
}

func TestNewTaxiiUserFail(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	_, err := newTaxiiUser("test@test.fail", "nopass")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiUserReadFail(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiUser.sql", "backend/sqlite/read/taxiiUser.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiUser.sql.testing", "backend/sqlite/read/taxiiUser.sql")

	_, err := newTaxiiUser("test@test.fail", "nopass")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiUserAssignedCollectionsReturnFail(t *testing.T) {
	defer setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = newTaxiiUser(testUser, testPass)
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestTaxiiUserAssignedCollectionsParseFail(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiCollectionAccess.sql", "backend/sqlite/read/taxiiCollectionAccess.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiCollectionAccess.sql.testing", "backend/sqlite/read/taxiiCollectionAccess.sql")

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = assignedCollections(s, "test@test.fail")
	if err == nil {
		t.Error("Expected an error")
	}
}