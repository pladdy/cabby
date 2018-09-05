package sqlite

import (
	"io/ioutil"
	"os"
	"strings"
	"time"

	"github.com/pladdy/cabby2/tester"
)

const (
	testDBPath = "testdata/tester.db"
	schema     = "schema.sql"
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

func createCollection(ds *DataStore, id string) {
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

	_, err = stmt.Exec(id, c.APIRootPath, c.Title, c.Description, strings.Join(c.MediaTypes, ","))
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

	_, err = stmt.Exec(tester.UserEmail, id, c.CanRead, c.CanWrite)
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

func createObject(ds *DataStore, id string) {
	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	stmt, err := tx.Prepare(`insert into stix_objects (id, type, created, modified, object, collection_id)
								           values (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	o := tester.Object
	_, err = stmt.Exec(id, o.Type, o.Created, o.Modified, string(o.Object), o.CollectionID.String())
	if err != nil {
		tester.Error.Fatal(err)
	}
	tx.Commit()
}

func createObjectVersion(ds *DataStore, id, version string) {
	tx, err := ds.DB.Begin()
	if err != nil {
		tester.Error.Fatal(err)
	}

	stmt, err := tx.Prepare(`insert into stix_objects (id, type, created, modified, object, collection_id)
								           values (?, ?, ?, ?, ?, ?)`)
	if err != nil {
		tester.Error.Fatal(err)
	}
	defer stmt.Close()

	o := tester.Object
	t := time.Now().UTC()

	_, err = stmt.Exec(id, o.Type, t.Format(time.RFC3339Nano), version, string(o.Object), o.CollectionID.String())
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

	tester.Info.Println("Setting up test sqlite db:", testDBPath)

	ds := testDataStore()

	f, err := os.Open(schema)
	if err != nil {
		tester.Error.Fatal("Couldn't open schema file: ", err)
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		tester.Error.Fatal("Couldn't read schema file: ", err)
	}

	_, err = ds.DB.Exec(string(schema))
	if err != nil {
		tester.Error.Fatal("Couldn't load schema: ", err)
	}

	createUser(ds)
	createDiscovery(ds)
	createAPIRoot(ds)
	createCollection(ds, tester.Collection.ID.String())
	createObject(ds, string(tester.Object.ID))
}

func testDataStore() *DataStore {
	ds, err := NewDataStore(testDBPath)
	if err != nil {
		tester.Error.Fatal(err)
	}
	return ds
}

func tearDownSQLite() {
	tester.Info.Println("Tearing down test sqlite db:", testDBPath)
	os.Remove(testDBPath)
}
