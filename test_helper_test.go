// test helper file to declare top level vars/constants and define helper functions for all tests

package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

var sqlDriver = "sqlite3"
var testDB = "test/test.db"

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		logError.Fatal("Failed to rename file: ", from, " to: ", to)
	}
}

func setupSQLite() {
	tearDownSQLite()

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		log.Fatal("Can't connect to test DB:", testDB)
	}

	f, err := os.Open("backend/sqlite/schema.sql")
	if err != nil {
		log.Fatal("Couldn't open schema file")
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Couldn't read schema file")
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		log.Fatal("Couldn't load schema")
	}
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
