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

func TestNewTaxiiID(t *testing.T) {
	// no string passed
	id, err := newTaxiiID()
	if err != nil {
		t.Error(err)
	}

	// empty string passed
	id, err = newTaxiiID("")
	if err != nil {
		t.Fatal(err)
	}

	if len(id.String()) == 0 {
		t.Error("Got:", id.String(), "Expected a taxiiID")
	}

	// uuid passed (valid taxiiID)
	uid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	id, err = newTaxiiID(uid)
	if err != nil {
		t.Fatal(err)
	}

	if id.String() != uid {
		t.Error("Got:", id.String(), "Expected:", uid)
	}

	// invalid uid passed
	id, err = newTaxiiID("fail")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestTaxiiCollectionCreate(t *testing.T) {
	setupSQLite()

	cid, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: cid, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(testUser, "api_root")
	if err != nil {
		t.Fatal(err)
	}

	// check on record
	time.Sleep(100 * time.Millisecond)

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// check id in taxii_collection
	var uid string
	err = s.db.QueryRow("select id from taxii_collection where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}

	// check for collection api root association
	err = s.db.QueryRow("select id from taxii_collection_api_root where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}
}

func TestTaxiiCollectionCreateFailTaxiiStorer(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}
	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiCollectionCreateFailQuery(t *testing.T) {
	renameFile("backend/sqlite/create/taxiiCollection.sql", "backend/sqlite/create/taxiiCollection.sql.testing")
	defer renameFile("backend/sqlite/create/taxiiCollection.sql.testing", "backend/sqlite/create/taxiiCollection.sql")

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionCreateFailWrite(t *testing.T) {
	defer setupSQLite()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionCreateFailWritePart(t *testing.T) {
	defer setupSQLite()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection_api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionRead(t *testing.T) {
	setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record and add a user to access it
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ("` + id.String() + `", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// buffered channel
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}
	results := make(chan interface{}, 10)
	go testCollection.read(testUser, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != id {
			t.Error("Got:", resultCollection.ID, "Expected", id.String())
		}
	}

	// unbuffered channel
	results = make(chan interface{})
	go testCollection.read(testUser, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != id {
			t.Error("Got:", resultCollection.ID, "Expected", id.String())
		}
	}
}

func TestTaxiiCollectionReadFail(t *testing.T) {
	defer reloadTestConfig()
	config = cabbyConfig{}
	config.DataStore = map[string]string{"name": "sqlite"}

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	results := make(chan interface{}, 10)
	go testCollection.read("user", results)

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

	reloadTestConfig()
	results = make(chan interface{}, 10)

	go testCollection.read("user", results)

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

	s, err := newSQLiteDB()
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

	user, err := newTaxiiUser(testUser, testPass)
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

	s, err := newSQLiteDB()
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

/* Config */

func TestParseConfig(t *testing.T) {
	config = cabbyConfig{}.parse("config/cabby.example.json")
	defer reloadTestConfig()

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
