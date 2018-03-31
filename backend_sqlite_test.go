package main

import (
	"io/ioutil"
	"strconv"
	"testing"
)

func TestNewSQLiteDB(t *testing.T) {
	s, err := newSQLiteDB(config.DataStore["path"])
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()
}

func TestNewSQLiteConnectFail(t *testing.T) {
	_, err := newSQLiteDB("")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestSQLiteConnectFailDriver(t *testing.T) {
	s, err := newSQLiteDB(config.DataStore["path"])
	if err != nil {
		t.Fatal(err)
	}

	s.driver = "sqlite"

	err = s.connect("doesn't matter")
	if err == nil {
		t.Error("Expected an error")
	}
}

/* sqlite parser */

func TestSQLiteParse(t *testing.T) {
	tests := []struct {
		command  string
		fileName string
		path     string
	}{
		{"create", "taxiiCollection", "backend/sqlite/create/taxiiCollection.sql"},
	}

	s := getSQLiteDB()
	defer s.disconnect()

	for _, test := range tests {
		result, err := s.parse(test.command, test.fileName)
		if err != nil {
			t.Fatal(err)
		}

		expected, err := ioutil.ReadFile(test.path)
		if err != nil {
			t.Fatal(err)
		}

		if result.query != string(expected) {
			t.Error("Got:", result.query, "Expected:", string(expected))
		}
	}
}

func TestSQLiteParseInvalid(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.parse("noCommand", "fail")
	if err == nil {
		t.Fatal("Expected error")
	}
}

/* sqlite reader functions */

func TestSQLiteRead(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	// create a collection record and add a user to access it
	tuid, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`insert into taxii_collection (id, api_root_path, title, description, media_types)
	                    values ("` + tuid.String() + `", "api_root", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ('` + testUser + `', "` + tuid.String() + `", 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	result, err := s.read("taxiiCollection", []interface{}{testUser, tuid.String()})
	if err != nil {
		t.Fatal(err)
	}

	tc := result.(taxiiCollections)

	if len(tc.Collections) == 0 {
		t.Error("Collections returned should be > 0")
	}
	if tc.Collections[0].ID.String() != tuid.String() {
		t.Error("Got:", tc.Collections[0].ID.String(), "Expected:", tuid.String())
	}
}

func TestSQLiteReadFail(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	result, err := s.read("taxiiCollection", []interface{}{"user1"})

	switch result := result.(type) {
	case error:
		err = result.(error)
	}
	if err == nil {
		t.Error("Expected an error")
	}
}

// new read* functions get tested here
func TestSQLiteReadScanError(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	tests := []struct {
		fn readFunction
	}{
		{s.readAPIRoot},
		{s.readAPIRoots},
		{s.readCollections},
		{s.readCollectionAccess},
		{s.readDiscovery},
		{s.readRoutableCollections},
		{s.readStixObjects},
		{s.readUser},
	}

	for _, test := range tests {
		rows, err := s.db.Query(`select 1,1,1,1,1,1,1,1,1,1,1,1,1,1,1`)
		if err != nil {
			t.Fatal(err)
		}
		_, err = test.fn(rows)

		if err == nil {
			t.Error("Expected error")
		}
	}
}

func TestReadRowsInvalid(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	rows, err := s.db.Query("select 1")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.readRows("invalid", rows)

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestSQLiteReadDiscoveryNoAPIRoot(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	rows, err := s.db.Query("select title, description, contact, default_url, 'test_api_root' from taxii_discovery")
	if err != nil {
		t.Fatal(err)
	}

	td, err := s.readDiscovery(rows)
	if err != nil {
		t.Fatal(err)
	}
	discovery := td.(taxiiDiscovery)

	if len(discovery.APIRoots) == 0 {
		t.Error("Got:", discovery.APIRoots, "Expected a non-empty API Roots")
	}
}

/* sqlite creator interface */

func TestSQLiteCreate(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go s.create("taxiiCollection", toCreate, errs)
	toCreate <- []interface{}{"test", "test api root", "test collection", "this is a test collection", "media type"}
	close(toCreate)

	for e := range errs {
		t.Fatal(e)
	}

	var uid string
	err := s.db.QueryRow(`select id from taxii_collection where id = 'test'`).Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "test" {
		t.Error("Got:", uid, "Expected:", "test")
	}
}

func TestSQLiteCreateFail(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)
	go s.create("invalidQuery", toCreate, errs)

	var err error
	for e := range errs {
		err = e
	}

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSQLiteCreateFailTransaction(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	toCreate := make(chan interface{}, 10)
	defer close(toCreate)
	errs := make(chan error, 10)
	go s.create("invalidQuery", toCreate, errs)

	var err error
	for e := range errs {
		err = e
	}

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSQLiteCreateFailExec(t *testing.T) {
	s := getSQLiteDB()
	defer s.disconnect()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go s.create("taxiiCollection", toCreate, errs)
	toCreate <- []interface{}{"not enough params"}
	close(toCreate)

	var err error
	for e := range errs {
		err = e
	}

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSQLiteCreateMaxWrites(t *testing.T) {
	setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	toCreate := make(chan interface{}, 10)
	defer close(toCreate)
	errs := make(chan error, 10)

	go s.create("taxiiCollection", toCreate, errs)

	// check errors while testing
	go func(t *testing.T) {
		for e := range errs {
			fail.Println("Should not have an error during test.  Error:", e)
			close(toCreate)
			t.Fatal(e)
		}
	}(t)

	writes := maxWrites
	for i := 0; i < writes; i++ {
		iStr := strconv.FormatInt(int64(i), 10)
		args := []interface{}{"test" + iStr, "apiRoot" + iStr, t.Name(), "description", "media_type"}
		toCreate <- args
	}

	var collections int
	err := s.db.QueryRow("select count(*) from taxii_collection where title = '" + t.Name() + "'").Scan(&collections)
	if err != nil {
		t.Fatal(err)
	}

	if collections != writes {
		t.Error("Got:", collections, "Expected:", writes)
	}
}
