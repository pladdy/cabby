package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
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

type handlerTestFunction func(http.HandlerFunc, string, string, *bytes.Buffer) (int, string)

const (
	testAdminPath    = "admin"
	testAPIRootPath  = "cabby_test_root"
	testConfigPath   = "testdata/config/testing_config.json"
	testDB           = "testdata/test.db"
	testCollectionID = "82407036-edf9-4c75-9a56-e72697c53e99"
	testHost         = "https://localhost:1234/"
	testPass         = "test"
	testPort         = 1234
	testUser         = "test@cabby.com"
	discoveryURL     = "https://localhost:1234/taxii/"
	eightMB          = 8388608
)

var (
	// logging
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)

	// test globals
	testAdminURL            = testHost + testAdminPath + "/"
	testAdminAPIRootURL     = testHost + testAdminPath + "/" + "api_root"
	testAdminCollectionsURL = testHost + testAdminPath + "/" + "collections"
	testAPIRootURL          = testHost + testAPIRootPath + "/"
	testAPIRoot             = taxiiAPIRoot{Path: testAPIRootPath,
		Title:            "test api root",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}
	testCollectionURL = testHost + testAPIRootPath + "/collections/" + testCollectionID + "/"
	testDiscovery     = taxiiDiscovery{Title: "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     "https://localhost/taxii/"}
	testObjectsURL = testHost + testAPIRootPath + "/collections/" + testCollectionID + "/objects/"
	testStatusURL  = testHost + testAPIRootPath + "/status/"
)

// attemptHandlerTest attempts a handler test by trying it up to maxTries times
func attemptHandlerTest(hf http.HandlerFunc, method, u string, b *bytes.Buffer) (int, string) {
	status, body := handlerTest(hf, method, u, b)
	attempts := 3

	for i := 1; i <= attempts; i++ {
		if status == http.StatusOK {
			break
		}

		time.Sleep(100 * time.Millisecond)
		status, body = handlerTest(hf, method, u, b)
	}
	return status, body
}

func createAPIRoot(testStorer taxiiStorer) {
	err := testAPIRoot.create(testStorer)
	if err != nil {
		fail.Fatal(err)
	}
}

func createCollection(testStorer taxiiStorer, cid string) {
	id, err := taxiiIDFromString(cid)
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
	tu := taxiiUser{Email: testUser, CanAdmin: true}
	err := tu.create(testStorer, fmt.Sprintf("%x", sha256.Sum256([]byte(testPass))))
	if err != nil {
		fail.Fatal(err)
	}
}

func getStorer() taxiiStorer {
	ts, err := newTaxiiStorer(testConfig().DataStore["name"], testConfig().DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	return ts
}

func getSQLiteDB() *sqliteDB {
	s, err := newSQLiteDB(testConfig().DataStore["path"])
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
		req = withAuthContext(httptest.NewRequest(method, url, b))
	} else {
		req = withAuthContext(httptest.NewRequest(method, url, nil))
	}

	res := httptest.NewRecorder()
	h(res, req)

	body, _ := ioutil.ReadAll(res.Body)
	return res.Code, string(body)
}

func testConfig() config {
	return configs{}.parse(testConfigPath)["testing"]
}

func postBundle(u, bundlePath string) (string, error) {
	ts := getStorer()
	defer ts.disconnect()

	// post a bundle to the data store
	bundleFile, _ := os.Open(bundlePath)
	bundleContent, _ := ioutil.ReadAll(bundleFile)

	maxContent := int64(4096)
	b := bytes.NewBuffer(bundleContent)
	status, body := handlerTest(handleTaxiiObjects(ts, maxContent), "POST", u, b)

	if !statusOkay(status) {
		fail.Println("Failed to post bundle", "Status:", status, "Body:", string(body))
	}

	var returnedStatus taxiiStatus
	err := json.Unmarshal([]byte(body), &returnedStatus)
	if err != nil {
		return "", err
	}

	waitForCompletion(returnedStatus)
	return returnedStatus.ID.String(), err
}

func renameFile(from, to string) {
	err := os.Rename(from, to)
	if err != nil {
		fail.Fatal("Failed to rename file: ", from, " to: ", to)
	}
}

func setupSQLite() {
	tearDownSQLite()

	info.Println("Setting up a test sqlite db:", testDB)
	var sqlDriver = "sqlite3"

	db, err := sql.Open(sqlDriver, testDB)
	if err != nil {
		fail.Fatal("Can't connect to test DB: ", testDB, "Error: ", err)
	}

	f, err := os.Open("build/debian/var/cabby/schema.sql")
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

	c := testConfig()

	ts, err := newTaxiiStorer(c.DataStore["name"], c.DataStore["path"])
	if err != nil {
		fail.Fatal(err)
	}
	defer ts.disconnect()

	createDiscovery(ts)
	createAPIRoot(ts)
	createUser(ts)
	createCollection(ts, testCollectionID)
}

// slowly post some bundles, but post the last one after a pause; useful for testing added_after parameter
func slowlyPosLasttBundle() (tm time.Time) {
	pause := 250 * time.Millisecond

	for i := range []int{0, 1, 2} {
		info.Printf("posting bundle...%v\n", i)
		if i == 2 {
			tm = time.Now().In(time.UTC)
			time.Sleep(pause)
		}
		postBundle(objectsURL(), fmt.Sprintf("testdata/added_after_%v.json", i))
	}
	return
}

func statusOkay(status int) bool {
	if status == http.StatusOK || status == http.StatusAccepted {
		return true
	}
	return false
}

func tearDownSQLite() {
	os.Remove(testDB)
}

func testingContext() context.Context {
	tid, err := taxiiIDFromString(testCollectionID)
	if err != nil {
		log.Fatal(err)
	}

	return context.WithValue(context.Background(),
		userCollections,
		map[taxiiID]taxiiCollectionAccess{tid: taxiiCollectionAccess{ID: tid, CanRead: true, CanWrite: true}})
}

func waitForCompletion(status taxiiStatus) (err error) {
	ts := getStorer()
	defer ts.disconnect()

	attempts := 3
	pause := 100

	for i := 1; i <= attempts; i++ {
		status.read(ts)
		if status.Status != "complete" {
			info.Println("Waiting for status to be complete")
			time.Sleep(time.Duration(i*pause) * time.Millisecond)
		}
	}
	return
}

// create a context for the testUser and give it read/write access to the test collection
func withAuthContext(r *http.Request) *http.Request {
	ctx := context.WithValue(testingContext(), userName, testUser)
	ctx = context.WithValue(ctx, canAdmin, true)
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
