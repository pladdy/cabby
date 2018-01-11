package main

import (
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
	"testing"
)

func init() {
	setupSQLite()
}

/* taxiiDataStorer */

func TestTaxiiDataStore(t *testing.T) {
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	tds, err := newTaxiiDataStore(c)
	if err != nil {
		t.Error(err)
	}
	defer tds.disconnect()
}

/* sqlite connector interface */

func TestSQLiteConnect(t *testing.T) {
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()
}

func TestSQLiteConnectFail(t *testing.T) {
	c := cabbyConfig{}
	c.DataStore = map[string]string{"name": "sqlite"}

	_, err := newSQLiteDB(c)
	if err == nil {
		t.Error("Expected an error")
	}
}

/* sqlite reader interface */

func TestSQLiteRead(t *testing.T) {
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record and add a user to access it
	tuid := uuid.Must(uuid.NewV4())
	_, err = s.db.Exec(`insert into taxii_collection values ("` + tuid.String() + `", "a title", "a description")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection values ("user1", "` + tuid.String() + `", 1, 1, "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// buffered channel
	results := make(chan interface{}, 10)
	go s.read("taxii_collection", map[string]string{"user_id": "user1"}, results)

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
	go s.read("taxii_collection", map[string]string{"user_id": "user1"}, results)

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

func TestSQLiteReadFail(t *testing.T) {
	setupSQLite() // reset state

	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan interface{}, 10)
	go s.read("taxii_collection", map[string]string{"user_id": "user1"}, results)

Loop:
	for r := range results {
		switch r := r.(type) {
		case error:
			logError.Println(r)
			break Loop
		}
		t.Error("Expected error")
	}

	setupSQLite() // reset state
}

func TestSQLiteReadCollectionScanError(t *testing.T) {
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	rows, err := s.db.Query(`select "fail"`)
	if err != nil {
		t.Fatal(err)
	}

	results := make(chan interface{})
	go s.readCollection(rows, results)

	for r := range results {
		switch r := r.(type) {
		case error:
			logError.Println(r)
			err = r
		}
	}

	if err == nil {
		t.Error("Expected error")
	}
}

/* sqlite writer interface */

func TestSQLiteCreateCollection(t *testing.T) {
	c := cabbyConfig{}.parse(configPath)
	c.DataStore["path"] = testDB

	s, err := newSQLiteDB(c)
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	collectionMap := map[string]string{"id": "test", "title": "test collection", "description": "this is a test collection"}
	_ = s.create("taxii_collection", collectionMap)

	// check
	var uid string
	err = s.db.QueryRow(`select id from taxii_collection where id = 'test'`).Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "test" {
		t.Error("Got:", uid, "Expected:", "test")
	}

	// create fail by trying the same insert (primary key violation)
	err = s.create("taxii_collection", collectionMap)
	if err == nil {
		t.Error("Expected an error")
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
