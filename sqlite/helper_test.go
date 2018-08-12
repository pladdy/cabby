package sqlite

import (
	"database/sql"
	"io/ioutil"
	"os"
	"strings"

	"github.com/pladdy/cabby2/tester"
)

const (
	testDB = "testdata/tester.db"
	schema = "schema.sql"
)

/* helpers */

func createAPIRoot(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	stmt, err := tx.Prepare(`insert into taxii_api_root
		(id, api_root_path, title, description, versions, max_content_length) values (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	a := tester.APIRoot

	_, err = stmt.Exec("testID", a.Path, a.Title, a.Description, strings.Join(a.Versions, ","), a.MaxContentLength)
	if err != nil {
		tester.Error.Fatal(err)
	}
	tx.Commit()
}

func createCollection(ds *DataStore) {
	c := tester.Collection

	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	// collection
	stmt, err := tx.Prepare(`insert into taxii_collection (id, api_root_path, title, description, media_types)
	  values (?, ?, ?, ?, ?)`)
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(c.ID.String(), c.APIRootPath, c.Title, c.Description, strings.Join(c.MediaTypes, ","))
	if err != nil {
		tester.Error.Fatal(err)
	}

	// user collection
	stmt, err = tx.Prepare(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
		values (?, ?, ?, ?)`)
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(tester.UserEmail, c.ID.String(), c.CanRead, c.CanWrite)
	if err != nil {
		tester.Error.Fatal(err)
	}
	tx.Commit()
}

func createDiscovery(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	stmt, err := tx.Prepare("insert into taxii_discovery (title, description, contact, default_url) values (?, ?, ?, ?)")
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	d := tester.Discovery
	_, err = stmt.Exec(d.Title, d.Description, d.Contact, d.Default)
	if err != nil {
		tester.Error.Fatal(err)
	}
	tx.Commit()
}

func createUser(ds *DataStore) {
	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	// user
	stmt, err := tx.Prepare("insert into taxii_user (email, can_admin) values (?, ?)")
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	u := tester.User

	_, err = stmt.Exec(u.Email, u.CanAdmin)
	if err != nil {
		tester.Error.Fatal(err)
	}

	// password
	stmt, err = tx.Prepare("insert into taxii_user_pass (email, pass) values (?, ?)")
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	_, err = stmt.Exec(u.Email, hash(tester.UserPassword))
	if err != nil {
		tester.Error.Fatal(err)
	}

	tx.Commit()
}

func setupSQLite() {
	tearDownSQLite()

	tester.Info.Println("Setting up a test sqlite db:", testDB)
	var sqlDriver = "sqlite3"

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		tester.Error.Fatal("Can't connect to test DB: ", testDB, "Error: ", err)
	}

	f, err := os.Open(schema)
	if err != nil {
		tester.Error.Fatal("Couldn't open schema file: ", err)
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		tester.Error.Fatal("Couldn't read schema file: ", err)
	}

	_, err = db.Exec(string(schema))
	if err != nil {
		tester.Error.Fatal("Couldn't load schema: ", err)
	}

	ds := testDataStore()
	createDiscovery(ds)
	createAPIRoot(ds)
	createCollection(ds)
	createUser(ds)
}

func testDataStore() *DataStore {
	ds, err := NewDataStore(testDB)
	if err != nil {
		tester.Error.Fatal(err)
	}
	return ds
}

func tearDownSQLite() {
	os.Remove(testDB)
}
