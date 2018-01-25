package main

import (
	"database/sql"
	"io/ioutil"
	"strconv"
	"testing"
	"time"
)

func init() {
	setupSQLite()
}

func TestTaxiiStorer(t *testing.T) {
	tds, err := newTaxiiStorer()
	if err != nil {
		t.Error(err)
	}
	defer tds.disconnect()
}

/* sqlite connector interface */

func TestSQLiteConnect(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()
}

func TestSQLiteConnectFail(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	config.DataStore = map[string]string{"name": "sqlite"}

	_, err := newSQLiteDB()
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

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
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

		if result != string(expected) {
			t.Error("Got:", result, "Expected:", string(expected))
		}
	}
}

func TestSQLiteParseInvalid(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.parse("noCommand", "fail")
	if err == nil {
		t.Fatal("Expected error")
	}
}

/* sqlite reader functions */

func TestSQLiteRead(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	// create a collection record and add a user to access it
	tuid, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ("` + tuid.String() + `", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ('` + testUser + `', "` + tuid.String() + `", 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	// buffered channel
	results := make(chan interface{}, 10)
	query, _ := s.parse("read", "taxiiCollection")
	go s.read(query, "taxiiCollection", []interface{}{testUser, tuid.String()}, results)

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
	go s.read(query, "taxiiCollection", []interface{}{testUser, tuid.String()}, results)

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

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	query, _ := s.parse("read", "taxiiCollection")
	results := make(chan interface{}, 10)
	go s.read(query, "taxiiCollection", []interface{}{"user1"}, results)

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

// new readX functions get tested here
func TestSQLiteReadScanError(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	type readFunction func(*sql.Rows, chan interface{})

	tests := []struct {
		fn readFunction
	}{
		{s.readCollection},
		{s.readUser},
	}

	for _, rf := range tests {
		rows, err := s.db.Query(`select "fail"`)
		if err != nil {
			t.Fatal(err)
		}

		results := make(chan interface{})
		go rf.fn(rows, results)

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
}

/* sqlite writer interface */

func TestSQLiteWrite(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	args := []interface{}{"test", "test collection", "this is a test collection", "media type"}
	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	query, err := s.parse("create", "taxiiCollection")
	if err != nil {
		t.Fatal(err)
	}

	go s.write(query, toWrite, errs)
	toWrite <- args
	close(toWrite)

	for e := range errs {
		t.Fatal(e)
	}

	var uid string
	err = s.db.QueryRow(`select id from taxii_collection where id = 'test'`).Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != "test" {
		t.Error("Got:", uid, "Expected:", "test")
	}
}

func TestSQLiteWriteFail(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)
	go s.write("invalidName", toWrite, errs)

	// sleep to let it fail in the go routine
	time.Sleep(100 * time.Millisecond)
}

func TestSQLiteWriteFailTransaction(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	toWrite := make(chan interface{}, 10)
	defer close(toWrite)
	errs := make(chan error, 10)
	go s.write("invalidName", toWrite, errs)

	for e := range errs {
		err = e
	}

	if err == nil {
		t.Fatal("Expected error")
	}
}

func TestSQLiteWriteFailExec(t *testing.T) {
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	args := []interface{}{"not enough params"}
	toWrite := make(chan interface{}, 10)
	defer close(toWrite)
	errs := make(chan error, 10)

	query, err := s.parse("create", "taxiiCollection")
	if err != nil {
		t.Fatal(err)
	}

	go s.write(query, toWrite, errs)
	toWrite <- args

	// sleep to let it fail in the go routine
	time.Sleep(100 * time.Millisecond)
}

func TestSQLiteWriteMaxWrites(t *testing.T) {
	setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	toWrite := make(chan interface{}, 10)
	defer close(toWrite)
	errs := make(chan error, 10)

	query, err := s.parse("create", "taxiiCollection")
	if err != nil {
		t.Fatal(err)
	}

	go s.write(query, toWrite, errs)

	// check errors while testing
	go func(t *testing.T) {
		for e := range errs {
			logError.Println("Should not have an error during test.  Error:", e)
			close(toWrite)
			t.Fatal(e)
		}
	}(t)

	writes := 500
	for i := 0; i < writes; i++ {
		iStr := strconv.FormatInt(int64(i), 10)
		args := []interface{}{"test" + iStr, t.Name(), "description", "media_type"}
		toWrite <- args
	}

	var collections int
	err = s.db.QueryRow("select count(*) from taxii_collection where title = '" + t.Name() + "'").Scan(&collections)
	if err != nil {
		t.Fatal(err)
	}

	if collections != writes {
		t.Error("Got:", collections, "Expected:", writes)
	}
}
