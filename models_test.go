package main

import (
	"io/ioutil"
	"log"
	"os"
	"strconv"
	"testing"

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
		title       string
		description string
	}{
		{"title1", "description1"},
		{"title2", "description2"},
	}

	// no uuid provided
	for _, test := range tests {
		c, err := newTaxiiCollection(map[string]string{"title": test.title, "description": test.description})
		if err != nil {
			t.Fatal(err)
		}

		if c.Title != test.title {
			t.Error("Got:", c.Title, "Expected:", test.title)
		}
		if c.Description != test.description {
			t.Error("Got:", c.Description, "Expected:", test.description)
		}
		if len(c.ID) == 0 {
			t.Error("UUID is empty")
		}
	}

	// uuid provided
	for _, test := range tests {
		tuid := uuid.Must(uuid.NewV4())
		c, err := newTaxiiCollection(map[string]string{"id": tuid.String(), "title": test.title, "description": test.description})
		if err != nil {
			t.Fatal(err)
		}

		if c.Title != test.title {
			t.Error("Got:", c.Title, "Expected:", test.title)
		}
		if c.Description != test.description {
			t.Error("Got:", c.Description, "Expected:", test.description)
		}
		if c.ID != tuid {
			t.Error("Got:", c.ID, "Expected:", tuid)
		}
	}
}

func TestNewCollectionFail(t *testing.T) {
	_, err := newTaxiiCollection(map[string]string{"id": "fail", "title": "a title", "description": "a description"})
	if err == nil {
		t.Error("Expected an error")
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
	defer tearDownSQLite()

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	cid := uuid.Must(uuid.NewV4())
	testCollection := taxiiCollection{ID: cid, Title: "test collection", Description: "a test collection"}

	err := testCollection.create(c)
	if err != nil {
		t.Error(err)
	}

	// check on record
	s, err := newSQLiteDB(c)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	var uid string
	err = s.db.QueryRow("select id from taxii_collection where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Error(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}
}

func TestTaxiiCollectionCreateFail(t *testing.T) {
	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	c := cabbyConfig{}
	err := testCollection.create(c)
	if err == nil {
		t.Error("Expected an error")
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
	user := "user1"

	_, err = s.db.Exec(`insert into taxii_collection values ("` + tuid.String() + `", "a title", "a description")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection values ("` + user + `", "` + tuid.String() + `", 1, 1, "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// buffered channel
	testCollection := taxiiCollection{ID: tuid, Title: "test collection", Description: "a test collection"}
	results := make(chan interface{}, 10)
	go testCollection.read(c, user, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != tuid {
			t.Error("Got:", resultCollection.ID, "Expected", "collection id")
		}
	}

	// unbuffered channel
	results = make(chan interface{})
	go testCollection.read(c, user, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			t.Fatal(r)
		}
		resultCollection := r.(taxiiCollection)

		if resultCollection.ID != tuid {
			t.Error("Got:", resultCollection.ID, "Expected", "collection id")
		}
	}
}

func TestTaxiiCollectionReadFail(t *testing.T) {
	c := cabbyConfig{}
	c.DataStore = map[string]string{"name": "sqlite"}

	testCollection := taxiiCollection{ID: uuid.Must(uuid.NewV4()), Title: "test collection", Description: "a test collection"}
	results := make(chan interface{}, 10)
	go testCollection.read(c, "user", results)

Loop:
	for r := range results {
		switch r := r.(type) {
		case error:
			logError.Println(r)
			break Loop
		}
		t.Error("Expected error")
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
