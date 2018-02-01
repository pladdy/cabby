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
	testUser     = "test@cabby.com"
	testPass     = "test"
	discoveryURL = "https://localhost:1234/taxii/"
	apiRootURL   = "https://localhost:1234/api_root/"
)

var testDB = "test/test.db"

func init() {
	setupSQLite()
}

func reloadTestConfig() {
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

	// create a user
	_, err = db.Exec(`insert into taxii_user (email) values('` + testUser + `')`)
	if err != nil {
		fail.Fatal("Couldn't add user")
	}

	pass := fmt.Sprintf("%x", sha256.Sum256([]byte(testPass)))
	_, err = db.Exec(`insert into taxii_user_pass (email, pass) values('` + testUser + `', '` + pass + `')`)
	if err != nil {
		fail.Fatalf("Couldn't add password: %v", err)
	}

	// create a collection
	testID := "82407036-edf9-4c75-9a56-e72697c53e99"
	_, err = db.Exec(`insert into taxii_collection (id, title, description, media_types)
											values ('` + testID + `', "a title", "a description", "")`)
	if err != nil {
		fail.Fatal("DB Err:", err)
	}

	// associate user to collection
	_, err = db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
											values ('` + testUser + `', '` + testID + `', 1, 1)`)
	if err != nil {
		fail.Fatal("DB Err:", err)
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
