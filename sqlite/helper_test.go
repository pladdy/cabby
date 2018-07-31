package sqlite

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"

	cabby "github.com/pladdy/cabby2"
)

const (
	testDB = "testdata/test.db"
	schema = "schema.sql"
)

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)

	testDiscovery = cabby.Discovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
)

func createDiscovery(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		fail.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into taxii_discovery (title, description, contact, default_url) values (?, ?, ?, ?)")
	if err != nil {
		fail.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(testDiscovery.Title, testDiscovery.Description, testDiscovery.Contact, testDiscovery.Default)
	if err != nil {
		fail.Fatal(err)
	}
	tx.Commit()
}

func testDataStore() *DataStore {
	ds, err := NewDataStore(testDB)
	if err != nil {
		fail.Fatal(err)
	}
	return ds
}

func setupSQLite() {
	tearDownSQLite()

	info.Println("Setting up a test sqlite db:", testDB)
	var sqlDriver = "sqlite3"

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		fail.Fatal("Can't connect to test DB: ", testDB, "Error: ", err)
	}

	f, err := os.Open(schema)
	if err != nil {
		fail.Fatal("Couldn't open schema file: ", err)
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		fail.Fatal("Couldn't read schema file: ", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		fail.Fatal("Couldn't load schema: ", err)
	}

	ds := testDataStore()
	createDiscovery(ds)
}

func tearDownSQLite() {
	os.Remove(testDB)
}
