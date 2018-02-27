package main

import (
	"bytes"
	"context"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestHandleTaxiiCollectionsPost(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", u.String(), b)

	if status != 200 {
		t.Error("Got:", status, "Expected: 200")
	}

	s := getSQLiteDB()
	defer s.disconnect()

	var title string
	err = s.db.QueryRow("select title from taxii_collection where title = '" + t.Name() + "'").Scan(&title)
	if err != nil {
		t.Error(err)
	}

	if title != t.Name() {
		t.Error("Got:", title, "Expected:", t.Name())
	}
}

func TestHandleTaxiiCollectionsPostCreateFail(t *testing.T) {
	defer setupSQLite()

	// remove required table
	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()

	// test a post which should fail
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Fatal(err)
	}

	b := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", u.String(), b)

	if status != 400 {
		t.Error("Got:", status, "Expected: 400")
	}
}

func TestHandleTaxiiCollectionsPostBadID(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", u.String(), nil)

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}

	// verify no record exists
	s := getSQLiteDB()
	defer s.disconnect()

	var title string
	err = s.db.QueryRow("select id from taxii_collection where id = 'fail'").Scan(&title)
	if err == nil {
		t.Fatal("Should be no record created")
	}
}

func TestHandleTaxiiCollectionsMethods(t *testing.T) {
	tests := []struct {
		method   string
		expected int
	}{
		{"POST", 400},
		{"CUSTOM", 405},
	}

	ts := getStorer()
	defer ts.disconnect()

	for _, test := range tests {
		req := httptest.NewRequest(test.method, "https://localhost/api_root/collections", nil)
		res := httptest.NewRecorder()
		h := handleTaxiiCollections(ts)
		h(res, req)

		if res.Code != test.expected {
			t.Error("Got:", res.Code, "Expected:", test.expected, "for method:", test.method)
		}
	}
}

func TestHandleTaxiiCollectionsPostNoUser(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections/")
	if err != nil {
		t.Fatal(err)
	}

	// set up a request with some data
	data := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	req := httptest.NewRequest("POST", u.String(), data)

	// update context to a fake user
	ctx := context.WithValue(context.Background(), userName, nil)
	req = req.WithContext(ctx)

	// record response
	res := httptest.NewRecorder()
	h := handleTaxiiCollections(ts)
	h(res, req)

	byteBody, _ := ioutil.ReadAll(res.Body)
	status, body := res.Code, string(byteBody)

	expected := `{"title":"Bad Request","description":"No user specified","http_status":"400"}` + "\n"

	if status != 400 {
		t.Error("Got:", status, "Expected: 400")
	}
	if body != expected {
		t.Error("Got:", body, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGet(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections/")
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", u.String(), nil)

	expected := `{"collections":[{"id":"82407036-edf9-4c75-9a56-e72697c53e99","can_read":true,` +
		`"can_write":true,"title":"a title","description":"a description","media_types":[""]}]}`

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetBadID(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections/fail")
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", u.String(), nil)
	expected := `{"title":"Bad Request","description":"uuid: incorrect UUID length: fail","http_status":"400"}` + "\n"

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetUnknownID(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", u.String(), nil)
	expected := `{"title":"Resource not found","description":"Invalid Collection","http_status":"404"}` + "\n"

	if status != 404 {
		t.Error("Got:", status, "Expected:", 404)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetInvalidUser(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("GET", u.String(), nil)
	// don't set a user in the context before setting up a response
	res := httptest.NewRecorder()
	h := handleTaxiiCollections(ts)
	h(res, req)

	status := res.Code

	if status != http.StatusUnauthorized {
		t.Error("Got:", status, "Expected:", http.StatusUnauthorized)
	}
}

func TestHandleTaxiiCollectionsGetReadError(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	s := getSQLiteDB()
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections(ts), "GET", u.String(), nil)
	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
}

func TestHandleTaxiiCollectionsGetNoResults(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	u, err := url.Parse("https://localhost/api_root/collections/")
	if err != nil {
		t.Fatal(err)
	}

	s := getSQLiteDB()
	defer s.disconnect()

	_, err = s.db.Exec("delete from taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections(ts), "GET", u.String(), nil)
	if status != 404 {
		t.Error("Got:", status, "Expected:", 404)
	}
}

func TestReadTaxiiCollection(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	w := httptest.NewRecorder()

	readTaxiiCollection(ts, w, testID, testUser)

	if w.Code != 200 {
		t.Error("Got:", w.Code, "Expected: 200")
	}
}

func TestReadTaxiiCollectionsFailRead(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	w := httptest.NewRecorder()
	user := "foo"

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	readTaxiiCollections(ts, w, user)

	if w.Code != 400 {
		t.Error("Got:", w.Code, "Expected: 404")
	}
}

/* model tests */

func TestNewCollection(t *testing.T) {
	tests := []struct {
		id          string
		shouldError bool
	}{
		{"invalid", true},
		{uuid.Must(uuid.NewV4()).String(), false},
	}

	// no uuid provided
	for _, test := range tests {
		c, err := newTaxiiCollection(test.id)

		if test.shouldError {
			if err == nil {
				t.Error("Test with id of", test.id, "should produce an error!")
			}
		} else if c.ID.String() != test.id {
			t.Error("Got:", c.ID.String(), "Expected:", test.id)
		}
	}
}

func TestNewTaxiiID(t *testing.T) {
	// no string passed
	id, err := newTaxiiID()
	if err != nil {
		t.Error(err)
	}

	// empty string passed
	id, err = newTaxiiID("")
	if err != nil {
		t.Fatal(err)
	}

	if len(id.String()) == 0 {
		t.Error("Got:", id.String(), "Expected a taxiiID")
	}

	// uuid passed (valid taxiiID)
	uid := "6ba7b810-9dad-11d1-80b4-00c04fd430c8"
	id, err = newTaxiiID(uid)
	if err != nil {
		t.Fatal(err)
	}

	if id.String() != uid {
		t.Error("Got:", id.String(), "Expected:", uid)
	}

	// invalid uid passed
	id, err = newTaxiiID("fail")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestTaxiiCollectionCreate(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	cid, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: cid, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(ts, testUser, "api_root")
	if err != nil {
		t.Fatal(err)
	}

	s := getSQLiteDB()
	defer s.disconnect()

	// check id in taxii_collection
	var uid string
	err = s.db.QueryRow("select id from taxii_collection where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}

	// check for collection api root association
	info.Println("Looking for ID:", cid.String())

	err = s.db.QueryRow(`select collection_id from taxii_collection_api_root where collection_id = '` +
		cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}
}

func TestTaxiiCollectionCreateFailQuery(t *testing.T) {
	renameFile("backend/sqlite/create/taxiiCollection.sql", "backend/sqlite/create/taxiiCollection.sql.testing")
	defer renameFile("backend/sqlite/create/taxiiCollection.sql.testing", "backend/sqlite/create/taxiiCollection.sql")

	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(ts, testUser, "api_root")
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionCreateFailWrite(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s := getSQLiteDB()
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(ts, testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionCreateFailWritePart(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s := getSQLiteDB()
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection_api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(ts, testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionRead(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	// create a collection record and add a user to access it
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ("` + id.String() + `", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ("` + testUser + `", "` + id.String() + `", 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	tc := taxiiCollection{ID: id}
	err = tc.read(ts, testUser)

	if tc.Title != "a title" {
		t.Error("Got:", tc.Title, "Expected", "a title")
	}
}

func TestTaxiiCollectionsRead(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	// create a collection record and add a user to access it
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.db.Exec(`insert into taxii_collection (id, title, description, media_types)
	                    values ("` + id.String() + `", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	tcs := taxiiCollections{}
	err = tcs.read(ts, testUser)
	if err != nil {
		t.Fatal(err)
	}

	if len(tcs.Collections) != 1 {
		t.Error("Got:", len(tcs.Collections), "Expected", 1)
	}
}

func TestTaxiiCollectionsReadFailQuery(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiCollections.sql", "backend/sqlite/read/taxiiCollections.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiCollections.sql.testing", "backend/sqlite/read/taxiiCollections.sql")

	ts := getStorer()
	defer ts.disconnect()

	tcs := taxiiCollections{}
	err := tcs.read(ts, testUser)
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionsReadFailWrite(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	tcs := taxiiCollections{}

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = tcs.read(ts, testUser)
	if err == nil {
		t.Error("Expected a write error")
	}
}
