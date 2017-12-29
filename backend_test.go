package main

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"testing"
)

/* helpers */

func setupSQLite() {
	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		log.Fatal("Can't connect to test DB:", testDB)
	}

	f, err := os.Open("backend/sql/schema.sql")
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
	err := os.Remove(testDB)
	if err != nil {
		log.Fatal("Failed to remove test db")
	}
}

/* test backend helpers */

func TestParseStatement(t *testing.T) {
	tests := []struct {
		language  string
		command   string
		container string
		path      string
	}{
		{"sql", "create", "taxii_collection", "backend/sql/create_taxii_collection.sql"},
	}

	for _, test := range tests {
		result := parseStatement(test.language, test.command, test.container)
		expected, err := ioutil.ReadFile(test.path)
		if err != nil {
			t.Error(err)
		}

		if result != string(expected) {
			t.Error("Got:", result, "Expected:", string(expected))
		}
	}
}

func TestParseStatementInvalid(t *testing.T) {
	p := panicChecker{recovered: false}
	defer attemptRecover(t, &p)

	parseStatement("nolang", "fail", "box")

	if p.recovered == false {
		t.Error("Expected recovered to be true")
	}
}

func TestSwapArgs(t *testing.T) {
	tests := []struct {
		statement string
		args      map[string]string
		expected  string
	}{
		{"insert into foo values('$id', '$title')",
			map[string]string{"id": "testId", "title": "test title"},
			"insert into foo values('testId', 'test title')",
		},
		{"delete foo",
			map[string]string{"id": "testId", "title": "test title"},
			"delete foo",
		},
	}

	for _, test := range tests {
		result := swapArgs(test.statement, test.args)
		if result != test.expected {
			t.Error("Got:", result, "Expected:", test.expected)
		}
	}
}

/* sqlite connector interface */

func TestSQLiteConnect(t *testing.T) {
	s := newSQLiteDB()

	err := s.connect(testDB)
	defer s.disconnect()

	if err != nil {
		t.Error(err)
	}
}

/* sqlite writer interface */

func TestSQLiteCreate(t *testing.T) {
	setupSQLite()
	defer tearDownSQLite()

	s := newSQLiteDB()

	err := s.connect(testDB)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	collection := map[string]string{"id": "test", "title": "test collection", "description": "this is a test collection"}
	s.create("taxii_collection", collection)

	// check
	var uid string
	err = s.db.QueryRow(`select id from taxii_collection where id = 'test'`).Scan(&uid)
	if err != nil {
		t.Error(err)
	}

	if uid != "test" {
		t.Error("Got:", uid, "Expected:", "test")
	}

	// create fail by trying the same insert (primary key violation)
	err = s.create("taxii_collection", collection)
	if err == nil {
		t.Error("Expected an error")
	}
}
