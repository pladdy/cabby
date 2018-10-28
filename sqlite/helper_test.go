package sqlite

import (
	"context"
	"io/ioutil"
	"os"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	testDBPath = "testdata/tester.db"
	schema     = "schema.sql"
)

func init() {
	// reduce logging spam during testing
	log.SetLevel(log.WarnLevel)
}

/* helpers */

func createAPIRoot(ds *DataStore) {
	err := ds.APIRootService().CreateAPIRoot(context.Background(), tester.APIRoot)
	if err != nil {
		log.Fatal(err)
	}
}

func createCollection(ds *DataStore, id string) {
	cid, _ := cabby.IDFromString(id)
	c := tester.Collection
	c.ID = cid

	err := ds.CollectionService().CreateCollection(context.Background(), c)
	if err != nil {
		log.Fatal(err)
	}

	ca := cabby.CollectionAccess{ID: c.ID, CanRead: true, CanWrite: true}
	err = ds.UserService().CreateUserCollection(context.Background(), tester.UserEmail, ca)
	if err != nil {
		log.Fatal(err)
	}
}

func createDiscovery(ds *DataStore) {
	err := ds.DiscoveryService().CreateDiscovery(context.Background(), tester.Discovery)
	if err != nil {
		log.Fatal(err)
	}
}

func createObject(ds *DataStore, id string) {
	o := tester.Object
	sid, _ := stones.IdentifierFromString(id)
	o.ID = sid

	err := ds.ObjectService().CreateObject(context.Background(), tester.Collection.ID.String(), o)
	if err != nil {
		log.Fatal(err)
	}
}

func createObjectVersion(ds *DataStore, id, version string) {
	o := tester.Object
	sid, _ := stones.IdentifierFromString(id)
	o.ID = sid
	o.Modified = version
	o.Created = time.Now().UTC().Format(time.RFC3339Nano)

	err := ds.ObjectService().CreateObject(context.Background(), tester.Collection.ID.String(), o)
	if err != nil {
		log.Fatal(err)
	}
}

func createUser(ds *DataStore) {
	err := ds.UserService().CreateUser(context.Background(), tester.User, tester.UserPassword)
	if err != nil {
		log.Fatal(err)
	}
}

func setupSQLite() {
	tearDownSQLite()
	log.Debug("Setting up test sqlite db:", testDBPath)

	f, err := os.Open(schema)
	if err != nil {
		log.Fatal("Couldn't open schema file: ", err)
	}

	schema, err := ioutil.ReadAll(f)
	if err != nil {
		log.Fatal("Couldn't read schema file: ", err)
	}

	ds := testDataStore()
	_, err = ds.DB.Exec(string(schema))
	if err != nil {
		log.Fatal("Couldn't load schema: ", err)
	}

	createUser(ds)
	createDiscovery(ds)
	createAPIRoot(ds)
	createCollection(ds, tester.Collection.ID.String())
	createObject(ds, string(tester.Object.ID.String()))
}

func testDataStore() *DataStore {
	ds, err := NewDataStore(testDBPath)
	if err != nil {
		log.Fatal(err)
	}
	return ds
}

func tearDownSQLite() {
	log.Debug("Tearing down test sqlite db:", testDBPath)
	os.Remove(testDBPath)
}
