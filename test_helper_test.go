package main

import (
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/ioutil"
	"os"
	"testing"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testAPIRootPath = "cabby_test_root"
	testConfig      = "test/config/testing_config.json"
	testDB          = "test/test.db"
	testID          = "82407036-edf9-4c75-9a56-e72697c53e99"
	testUser        = "test@cabby.com"
	testPass        = "test"
	discoveryURL    = "https://localhost:1234/taxii/"
	eightMB         = 8388608
)

var (
	testAPIRootURL = "https://localhost:1234/" + testAPIRootPath + "/"
	testAPIRoot    = taxiiAPIRoot{Title: "test api root",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
	testDiscovery = taxiiDiscovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
)

func createAPIRoot(testStorer taxiiStorer) {
	err := testAPIRoot.create(testStorer, testAPIRootPath)
	if err != nil {
		fail.Fatal(err)
	}
}

func createCollection(testStorer taxiiStorer) {
	id, err := newTaxiiID(testID)
	if err != nil {
		fail.Fatal(err)
	}

	tc := taxiiCollection{ID: id, Title: "a title", Description: "a description"}
	err = tc.create(testStorer, testUser, testAPIRootPath)

	if err != nil {
		fail.Fatal("DB Err:", err)
	}
}

func createDiscovery(testStorer taxiiStorer) {
	err := testDiscovery.create(testStorer)
	if err != nil {
		fail.Fatal(err)
	}
}

func createUser(testStorer taxiiStorer) {
	tu := taxiiUser{Email: testUser}
	err := tu.create(testStorer, fmt.Sprintf("%x", sha256.Sum256([]byte(testPass))))
	if err != nil {
		fail.Fatal(err)
	}
}

func getStorer() taxiiStorer {
	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	return ts
}

func getSQLiteDB() *sqliteDB {
	s, err := newSQLiteDB(config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	return s
}

func loadTestConfig() {
	config = cabbyConfig{}.parse(testConfig)
}

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		fail.Fatal("Failed to rename file: ", from, " to: ", to)
	}
}

func setupSQLite() {
	tearDownSQLite()

	info.Println("Setting up a test sqlite db:", testDB)
	var sqlDriver = "sqlite3"

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		fail.Fatal("Can't connect to test DB: ", testDB, "Error: ", err)
	}

	info.Println("Reading in schema")
	f, err := os.Open("backend/sqlite/schema.sql")
	if err != nil {
		fail.Fatal("Couldn't open schema file")
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		fail.Fatal("Couldn't read schema file")
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		fail.Fatal("Couldn't load schema")
	}

	loadTestConfig()

	info.Println("Creating resources")
	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	defer ts.disconnect()

	createDiscovery(ts)
	createAPIRoot(ts)
	createUser(ts)
	createCollection(ts)
}

func tearDownSQLite() {
	os.Remove(testDB)
}

/* check for panics and record recovery */

type panicChecker struct {
	recovered bool
}

func attemptRecover(t *testing.T, p *panicChecker) {
	if err := recover(); err == nil {
		t.Error("Failed to recover:", err)
	}
	p.recovered = true
}
