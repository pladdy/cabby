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

	// check
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
