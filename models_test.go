package main

import (
	"crypto/sha256"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"
	"time"

	uuid "github.com/satori/go.uuid"
)

/* API Roots */

func TestAPIRootVerify(t *testing.T) {
	tests := []struct {
		apiRoot      string
		rootMapEntry string
		expected     bool
	}{
		{"https://localhost/api_test", "https://localhost/api_test", true},
		{"https://localhost/api_fail", "https://localhost/api_test", false},
	}

	for _, test := range tests {
		// create a config struct with an API Root and corresponding API Root Map
		a := taxiiAPIRoot{
			Title:            "test",
			Description:      "test api root",
			Versions:         []string{"taxii-2.0"},
			MaxContentLength: 1}

		c := cabbyConfig{APIRootMap: map[string]taxiiAPIRoot{test.rootMapEntry: a}}

		c.Discovery = taxiiDiscovery{APIRoots: []string{test.apiRoot}}

		result := c.validAPIRoot(test.apiRoot)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

/* collections */

func TestNewCollection(t *testing.T) {
	tests := []struct {
		id          string
		shouldError bool
	}{
		{"invalid", true},
		{uuid.Must(uuid.NewV4()).String(), false},
	}

	// no uuid provided
	for _, test := range tests {
		c, err := newTaxiiCollection(test.id)

		if test.shouldError {
			if err == nil {
				t.Error("Test with id of", test.id, "should produce an error!")
			}
		} else if c.ID.String() != test.id {
			t.Error("Got:", c.ID.String(), "Expected:", test.id)
		}
	}
}

func TestNewUUID(t *testing.T) {
	// no string passed
	uid, err := newUUID()
	if err != nil {
		t.Fatal(err)
	}

	// empty string passed
	uid, err = newUUID("")
	if err != nil {
		t.Fatal(err)
	}

	if len(uid) == 0 {
		t.Error("Got:", uid.String(), "Expected a UUID")
	}

	// uuid string passed
	tuid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	uid, err = newUUID(tuid)
	if err != nil {
		t.Fatal(err)
	}

	if uid.String() != tuid {
		t.Error("Got:", uid.String(), "Expected:", tuid)
	}

	// invalid uid passed
	uid, err = newUUID("fail")
	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestTaxiiCollectionCreate(t *testing.T) {
	setupSQLite()

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	cid := uuid.Must(uuid.NewV4())
	testCollection := taxiiCollection{ID: cid, Title: "test collection", Description: "a test collection"}

	err := testCollection.create(c)
	if err != nil {
		t.Fatal(err)
	}

	// check on record
	time.Sleep(100 * time.Millisecond)

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	var uid string
	err = s.db.QueryRow("select id from taxii_collection where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}
}

func TestTaxiiCollectionCreateFailTaxiiStorer(t *testing.T) {
	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	c := cabbyConfig{}
	err := testCollection.create(c)
	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiCollectionCreateFailQuery(t *testing.T) {
	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	defer renameFile("backend/sqlite/create/taxiiCollection.sql.testing", "backend/sqlite/create/taxiiCollection.sql")
	renameFile("backend/sqlite/create/taxiiCollection.sql", "backend/sqlite/create/taxiiCollection.sql.testing")

	err := testCollection.create(c)
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionCreateFailWrite(t *testing.T) {
	defer setupSQLite()

	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(c)
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionRead(t *testing.T) {
	setupSQLite()

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record and add a user to access it
	tuid := uuid.Must(uuid.NewV4())
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ("` + tuid.String() + `", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// buffered channel
	testCollection := taxiiCollection{ID: tuid, Title: "test collection", Description: "a test collection"}
	results := make(chan interface{}, 10)
	go testCollection.read(c, testUser, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != tuid {
			t.Error("Got:", resultCollection.ID, "Expected", tuid.String())
		}
	}

	// unbuffered channel
	results = make(chan interface{})
	go testCollection.read(c, testUser, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != tuid {
			t.Error("Got:", resultCollection.ID, "Expected", tuid.String())
		}
	}
}

func TestTaxiiCollectionReadFail(t *testing.T) {
	// fail due to no valid database specified
	c := cabbyConfig{}
	c.DataStore = map[string]string{"name": "sqlite"}

	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	results := make(chan interface{}, 10)
	go testCollection.read(c, "user", results)

NoRecord:
	for r := range results {
		switch r := r.(type) {
		case error:
			logError.Println(r)
			break NoRecord
		}
		t.Error("Expected error")
	}

	// fail due to no query being avialable
	defer renameFile("backend/sqlite/read/taxiiCollection.sql.testing", "backend/sqlite/read/taxiiCollection.sql")
	renameFile("backend/sqlite/read/taxiiCollection.sql", "backend/sqlite/read/taxiiCollection.sql.testing")

	c = cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB
	results = make(chan interface{}, 10)

	go testCollection.read(c, "user", results)

NoQuery:
	for r := range results {
		switch r := r.(type) {
		case error:
			logError.Println(r)
			break NoQuery
		}
		t.Error("Expected error")
	}
}

/* taxiiUser */

func TestNewTaxiiUser(t *testing.T) {
	setupSQLite()

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record
	collectionID := uuid.Must(uuid.NewV4())
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ('` + collectionID.String() + `', "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// associate user to collection
	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ('` + testUser + `', '` + collectionID.String() + `', 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	user, err := newTaxiiUser(c, testUser, testPass)
	if err != nil {
		t.Fatal(err)
	}

	if user.Email != testUser {
		t.Error("Got:", user.Email, "Expected:", testUser)
	}

	for k, v := range user.CollectionAccess {
		logInfo.Println("Got collection id of", k)

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

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record
	collectionID := uuid.Must(uuid.NewV4())
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ('` + collectionID.String() + `', "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(testPass)))
	_, err = newTaxiiUser(c, testUser, pass)
	if err == nil {
		t.Error("Expected error with no access")
	}
}

func TestNewTaxiiUserFail(t *testing.T) {
	c := cabbyConfig{}
	_, err := newTaxiiUser(c, "test@test.fail", "nopass")
	if err == nil {
		t.Error("Expected an error")
	}

	// fail due to no query being avialable
	defer renameFile("backend/sqlite/read/taxiiUser.sql.testing", "backend/sqlite/read/taxiiUser.sql")
	renameFile("backend/sqlite/read/taxiiUser.sql", "backend/sqlite/read/taxiiUser.sql.testing")

	c = cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	_, err = newTaxiiUser(c, "test@test.fail", "nopass")
	if err == nil {
		t.Error("Expected an error")
	}

}

/* Config */

func TestParseConfig(t *testing.T) {
	config := cabbyConfig{}.parse("config/cabby.example.json")

	if config.Host != "localhost" {
		t.Error("Got:", "localhost", "Expected:", "localhost")
	}
	if config.Port != 1234 {
		t.Error("Got:", strconv.Itoa(1234), "Expected:", strconv.Itoa(1234))
	}
}

func TestParseConfigNotFound(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
		}
	}()

	_ = cabbyConfig{}.parse("foo/bar")
	t.Error("Failed to panic with an unknown resource")
}

func TestParseConfigInvalidJSON(t *testing.T) {
	invalidJSON := "invalid.json"

	defer func() {
		if r := recover(); r != nil {
			log.Println("Recovered", r)
			os.Remove(invalidJSON)
		}
	}()

	_ = ioutil.WriteFile(invalidJSON, []byte("invalid"), 0644)
	_ = cabbyConfig{}.parse(invalidJSON)
	t.Error("Failed to panic with an unknown resource")
}
