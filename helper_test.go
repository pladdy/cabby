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
	testID          = "82407036-edf9-4c75-9a56-e72697c53e99"
	testUser        = "test@cabby.com"
	testPass        = "test"
	discoveryURL    = "https://localhost:1234/taxii/"
	eightMB         = 8388608
)

var (
	apiRootURL  = "https://localhost:1234/" + testAPIRootPath + "/"
	testAPIRoot = taxiiAPIRoot{Title: "test api root",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
	testDB        = "test/test.db"
	testDiscovery = taxiiDiscovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
)

func init() {
	setupSQLite()
}

func createAPIRoot() {
	err := testAPIRoot.create(testAPIRootPath)
	if err != nil {
		fail.Fatal(err)
	}
}

func createCollection() {
	id, err := newTaxiiID(testID)
	if err != nil {
		fail.Fatal(err)
	}

	tc := taxiiCollection{ID: id, Title: "a title", Description: "a description"}
	err = tc.create(testUser, testAPIRootPath)

	if err != nil {
		fail.Fatal("DB Err:", err)
	}
}

func createDiscovery() {
	err := testDiscovery.create()
	if err != nil {
		fail.Fatal(err)
	}
}

func createUser() {
	tu := taxiiUser{Email: testUser}
	err := tu.create(fmt.Sprintf("%x", sha256.Sum256([]byte(testPass))))
	if err != nil {
		fail.Fatal(err)
	}
}

func loadTestConfig() {
	var testConfig = "test/config/testing_config.json"
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
	createDiscovery()
	createAPIRoot()
	createUser()
	createCollection()
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
