package main

import (
	"bytes"
	"context"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func collectionsURL() string {
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		fail.Fatal(err)
	}
	return u.String()
}

func TestHandleTaxiiCollectionsPost(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	b := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", collectionsURL(), b)

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}

	s := getSQLiteDB()
	defer s.disconnect()

	var title string
	err := s.db.QueryRow("select title from taxii_collection where title = '" + t.Name() + "'").Scan(&title)
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
	b := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", collectionsURL(), b)

	if status != http.StatusInternalServerError {
		t.Error("Got:", status, "Expected:", http.StatusInternalServerError)
	}
}

func TestHandleTaxiiCollectionsPostBadID(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiCollections(ts), "POST", collectionsURL(), nil)

	if status != http.StatusBadRequest {
		t.Error("Got:", status, "Expected:", http.StatusBadRequest)
	}

	// verify no record exists
	s := getSQLiteDB()
	defer s.disconnect()

	var title string
	err := s.db.QueryRow("select id from taxii_collection where id = 'fail'").Scan(&title)
	if err == nil {
		t.Fatal("Should be no record created")
	}
}

func TestHandleTaxiiCollectionsMethods(t *testing.T) {
	tests := []struct {
		method   string
		expected int
	}{
		{"POST", http.StatusBadRequest},
		{"CUSTOM", http.StatusMethodNotAllowed},
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

	// set up a request with some data
	data := bytes.NewBuffer([]byte(`{"title":"` + t.Name() + `"}`))
	req := httptest.NewRequest("POST", collectionsURL(), data)

	// update context to a fake user
	ctx := context.WithValue(context.Background(), userName, nil)
	req = req.WithContext(ctx)

	// record response
	res := httptest.NewRecorder()
	h := handleTaxiiCollections(ts)
	h(res, req)

	byteBody, _ := ioutil.ReadAll(res.Body)
	status, body := res.Code, string(byteBody)

	expected := `{"title":"Unauthorized","description":"No user specified","http_status":"401"}`

	if status != http.StatusUnauthorized {
		t.Error("Got:", status, "Expected:", http.StatusUnauthorized)
	}
	if body != expected {
		t.Error("Got:", body, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGet(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", collectionsURL(), nil)

	expected := `{"collections":[{"id":"82407036-edf9-4c75-9a56-e72697c53e99","can_read":true,` +
		`"can_write":true,"title":"a title","description":"a description","media_types":[""]}]}`

	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetBadID(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", collectionsURL()+"/fail", nil)
	expected := `{"title":"Resource not found","description":"uuid: incorrect UUID length: fail","http_status":"404"}`

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	// post an additional collection
	newID := "82407036-edf9-4c75-9a56-e72697c53e90"
	createCollection(ts, newID)

	// create request and add a range to it
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", collectionsURL(), nil))

	first := 0
	last := 0
	records := 1
	totalRecords := 2

	req.Header.Set("Range", "items "+strconv.Itoa(first)+"-"+strconv.Itoa(last))

	res := httptest.NewRecorder()
	handleTaxiiCollections(ts)(res, req)

	body, _ := ioutil.ReadAll(res.Body)

	var collections taxiiCollections
	err := json.Unmarshal([]byte(body), &collections)
	if err != nil {
		t.Fatal(err)
	}

	if res.Code != http.StatusPartialContent {
		t.Error("Got:", res.Code, "Expected:", http.StatusPartialContent)
	}

	if len(collections.Collections) != records {
		t.Error("Got:", len(collections.Collections), "Expected:", records)
	}

	tr := taxiiRange{first: int64(first), last: int64(last), total: int64(totalRecords)}
	if res.Header().Get("Content-Range") != tr.String() {
		t.Error("Got:", res.Header().Get("Content-Range"), "Expected:", tr.String())
	}
}

func TestHandleTaxiiCollectionsGetInvalidRange(t *testing.T) {
	setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	// create request and add a range to it that's invalid
	var req *http.Request
	req = withAuthContext(httptest.NewRequest("GET", collectionsURL(), nil))
	req.Header.Set("Range", "invalid range")

	res := httptest.NewRecorder()
	handleTaxiiCollections(ts)(res, req)

	if res.Code != http.StatusRequestedRangeNotSatisfiable {
		t.Error("Got:", res.Code, "Expected:", http.StatusRequestedRangeNotSatisfiable)
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

	status, result := handlerTest(handleTaxiiCollections(ts), "GET", collectionsURL()+"/"+id.String(), nil)
	expected := `{"title":"Resource not found","description":"Invalid collection","http_status":"404"}`

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
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

	req := httptest.NewRequest("GET", collectionsURL()+"/"+id.String(), nil)
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
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
}

func TestHandleTaxiiCollectionsGetNoResults(t *testing.T) {
	defer setupSQLite()

	ts := getStorer()
	defer ts.disconnect()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("delete from taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections(ts), "GET", collectionsURL(), nil)
	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
	}
}

func TestReadTaxiiCollection(t *testing.T) {
	ts := getStorer()
	defer ts.disconnect()

	status, _ := handlerTest(handleTaxiiCollections(ts), "GET", testCollectionURL, nil)
	if status != http.StatusOK {
		t.Error("Got:", status, "Expected:", http.StatusOK)
	}
}

func TestReadTaxiiCollectionsFailRead(t *testing.T) {
	defer setupSQLite()

	s := getSQLiteDB()
	defer s.disconnect()

	_, err := s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	ts := getStorer()
	defer ts.disconnect()
	status, _ := handlerTest(handleTaxiiCollections(ts), "GET", collectionsURL(), nil)

	if status != http.StatusNotFound {
		t.Error("Got:", status, "Expected:", http.StatusNotFound)
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

	_, err = s.db.Exec("drop table taxii_collection")
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
	_, err = s.db.Exec(`insert into taxii_collection (id, api_root_path, title, description, media_types)
	                    values ("` + id.String() + `", "api_root", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	_, err = s.db.Exec(`insert into taxii_user_collection (email, collection_id, can_read, can_write)
	                    values ("` + testUser + `", "` + id.String() + `", 1, 1)`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	tc := taxiiCollection{ID: id}
	_, err = tc.read(ts, testUser)

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
	_, err = s.db.Exec(`insert into taxii_collection (id, api_root_path, title, description, media_types)
	                    values ("` + id.String() + `", "api_root", "a title", "a description", "")`)
	if err != nil {
		t.Fatal("DB Err:", err)
	}

	tcs := taxiiCollections{}
	_, err = tcs.read(ts, testUser, taxiiFilter{})
	if err != nil {
		t.Fatal(err)
	}

	if len(tcs.Collections) != 1 {
		t.Error("Got:", len(tcs.Collections), "Expected", 1)
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

	_, err = tcs.read(ts, testUser, taxiiFilter{})
	if err == nil {
		t.Error("Expected a write error")
	}
}
