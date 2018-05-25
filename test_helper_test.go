package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	_ "github.com/mattn/go-sqlite3"
)

const (
	testAPIRootPath = "cabby_test_root"
	testConfig      = "testdata/config/testing_config.json"
	testDB          = "testdata/test.db"
	testID          = "82407036-edf9-4c75-9a56-e72697c53e99"
	testPass        = "test"
	testPort        = 1234
	testUser        = "test@cabby.com"
	discoveryURL    = "https://localhost:1234/taxii/"
	eightMB         = 8388608
)

var (
	// logging
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)

	// test globals
	testAPIRootURL = "https://localhost:1234/" + testAPIRootPath + "/"
	testAPIRoot    = taxiiAPIRoot{Title: "test api root",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
	testCollectionURL = "https://localhost:1234/" + testAPIRootPath + "/collections/" + testID + "/"
	testDiscovery     = taxiiDiscovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
	testObjectsURL = "https://localhost:1234/" + testAPIRootPath + "/collections/" + testID + "/objects/"
)

func createAPIRoot(testStorer taxiiStorer) {
	err := testAPIRoot.create(testStorer, testAPIRootPath)
	if err != nil {
		fail.Fatal(err)
	}
}

func createCollection(testStorer taxiiStorer, cid string) {
	id, err := newTaxiiID(cid)
	if err != nil {
		fail.Fatal(err)
	}

	tc := taxiiCollection{ID: id, Title: "a title", Description: "a description"}
	err = tc.create(testStorer, testUser, testAPIRootPath)

	if err != nil {
		fail.Fatal("DB Err:", err)
	}
}

func createDiscovery(testStorer taxiiStorer) {
	err := testDiscovery.create(testStorer)
	if err != nil {
		fail.Fatal(err)
	}
}

func createUser(testStorer taxiiStorer) {
	tu := taxiiUser{Email: testUser}
	err := tu.create(testStorer, fmt.Sprintf("%x", sha256.Sum256([]byte(testPass))))
	if err != nil {
		fail.Fatal(err)
	}
}

func getStorer() taxiiStorer {
	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	return ts
}

func getSQLiteDB() *sqliteDB {
	s, err := newSQLiteDB(config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	return s
}

// handle generic testing of handlers.  It takes a handler function to call with a url;
// it returns the status code and response as a string
func handlerTest(h http.HandlerFunc, method, url string, b *bytes.Buffer) (int, string) {
	var req *http.Request

	if b != nil {
		req = withAuthContext(httptest.NewRequest("POST", url, b))
	} else {
		req = withAuthContext(httptest.NewRequest(method, url, nil))
	}

	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

func loadTestConfig() {
	config = Config{}.parse(testConfig)
}

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		fail.Fatal("Failed to rename file: ", from, " to: ", to)
	}
}

func postBundle(u, bundlePath string) {
	ts := getStorer()
	defer ts.disconnect()

	// post a bundle to the data store
	bundleFile, _ := os.Open(bundlePath)
	bundleContent, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(2048)
	b := bytes.NewBuffer(bundleContent)
	handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	// give time for bundle to be persisted
	time.Sleep(100 * time.Millisecond)
}

func setupSQLite() {
	tearDownSQLite()

	info.Println("Setting up a test sqlite db:", testDB)
	var sqlDriver = "sqlite3"

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		fail.Fatal("Can't connect to test DB: ", testDB, "Error: ", err)
	}

	f, err := os.Open("backend/sqlite/schema.sql")
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

	loadTestConfig()

	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	defer ts.disconnect()

	createDiscovery(ts)
	createAPIRoot(ts)
	createUser(ts)
	createCollection(ts, testID)
}

// slowly post the malware bundle test files; used for pull back objects that are added after
// the returned time (which is just before the last bundle is posted to the server
func slowlyPostBundle() (tm time.Time) {
	for i := range []int{0, 1, 2} {
		info.Printf("posting bundle...%v\n", i)

		tm = time.Now().In(time.UTC)
		postBundle(objectsURL(), fmt.Sprintf("testdata/added_after_%v.json", i))
		time.Sleep(100 * time.Millisecond)
	}
	return
}

func tearDownSQLite() {
	os.Remove(testDB)
}

// create a context for the testUser and give it read/write access to the test collection
func withAuthContext(r *http.Request) *http.Request {
	tid, err := newTaxiiID(testID)
	if err != nil {
		log.Fatal(err)
	}

	ctx := context.WithValue(context.Background(), userName, testUser)
	ctx = context.WithValue(ctx,
		userCollections,
		map[taxiiID]taxiiCollectionAccess{tid: taxiiCollectionAccess{ID: tid, CanRead: true, CanWrite: true}})

	return r.WithContext(ctx)
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
