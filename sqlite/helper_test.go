package sqlite

import (
	"database/sql"
	"io/ioutil"
	"log"
	"os"
	"strings"

	cabby "github.com/pladdy/cabby2"
)

const (
	eightMB          = 8388608
	testAPIRootPath  = "cabby_test_root"
	testDB           = "testdata/test.db"
	testUserEmail    = "test@cabby.com"
	testUserPassword = "test"
	schema           = "schema.sql"
)

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)

	testAPIRoot = cabby.APIRoot{Path: testAPIRootPath,
		Title:            "test api root title",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
	testDiscovery = cabby.Discovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
	testUser = cabby.User{Email: testUserEmail,
		CanAdmin: true}
)

/* helpers */

func createAPIRoot(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		fail.Fatal(err)
	}

	stmt, err := tx.Prepare(`insert into taxii_api_root
		(id, api_root_path, title, description, versions, max_content_length) values (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		fail.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec("testID",
		testAPIRootPath, testAPIRoot.Title, testAPIRoot.Description, strings.Join(testAPIRoot.Versions, ","), testAPIRoot.MaxContentLength)
	if err != nil {
		fail.Fatal(err)
	}
	tx.Commit()
}

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

func createUser(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		fail.Fatal(err)
	}

	// user
	stmt, err := tx.Prepare("insert into taxii_user (email, can_admin) values (?, ?)")
	if err != nil {
		fail.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(testUser.Email, testUser.CanAdmin)
	if err != nil {
		fail.Fatal(err)
	}

	// password
	stmt, err = tx.Prepare("insert into taxii_user_pass (email, pass) values (?, ?)")
	if err != nil {
		fail.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(testUser.Email, hash(testUserPassword))
	if err != nil {
		fail.Fatal(err)
	}

	tx.Commit()
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
	createAPIRoot(ds)
	createUser(ds)
}

func tearDownSQLite() {
	os.Remove(testDB)
}

func testDataStore() *DataStore {
	ds, err := NewDataStore(testDB)
	if err != nil {
		fail.Fatal(err)
	}
	return ds
}
