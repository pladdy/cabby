package main

import (
	"io/ioutil"
	"net/http/httptest"
	"net/url"
	"testing"

	uuid "github.com/satori/go.uuid"
)

func TestHandleTaxiiCollectionsPost(t *testing.T) {
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Error(err)
	}

	q := u.Query()
	q.Set("title", t.Name())
	q.Set("description", "a description")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollections, "POST", u.String())

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
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

func TestHandleTaxiiCollectionsPostBadID(t *testing.T) {
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Error(err)
	}

	q := u.Query()
	q.Set("id", "fail")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollections, "POST", u.String())

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}

	// verify no record exists
	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	var title string
	err = s.db.QueryRow("select id from taxii_collection where id = 'fail'").Scan(&title)
	if err == nil {
		t.Fatal("Should be no record created")
	}
}

func TestHandleTaxiiCollectionsPostBadParse(t *testing.T) {
	tests := []struct {
		method   string
		expected int
	}{
		{"POST", 400},
		{"CUSTOM", 400},
	}

	for _, test := range tests {
		req := httptest.NewRequest(test.method, "https://localhost/api_root/collections", nil)

		// change body to nil to trigger a parse error in handler
		if test.method == "POST" {
			req.Body = nil
		}

		res := httptest.NewRecorder()
		handleTaxiiCollections(res, req)

		if res.Code != test.expected {
			t.Error("Got:", res.Code, "Expected:", test.expected, "for method:", test.method)
		}
	}
}

func TestHandleTaxiiCollectionsPostInvalidDB(t *testing.T) {
	config = cabbyConfig{}.parse("test/config/no_datastore_config.json")
	defer reloadTestConfig()

	// set up URL
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()
	q.Set("title", "a title")
	q.Set("description", "a description")
	u.RawQuery = q.Encode()

	status, _ := handlerTest(handleTaxiiCollections, "POST", u.String())

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
}

func TestHandleTaxiiCollectionsPostInvalidUser(t *testing.T) {
	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	req := httptest.NewRequest("POST", u.String(), nil)
	// don't set a user in the context before setting up a response
	res := httptest.NewRecorder()
	handleTaxiiCollections(res, req)
	b, _ := ioutil.ReadAll(res.Body)

	status, result := res.Code, string(b)
	expected := `{"title":"Bad Request","description":"Invalid user specified","http_status":"400"}` + "\n"

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetCollection(t *testing.T) {
	// create a collection
	u, err := url.Parse("https://localhost/api_root/collections")
	if err != nil {
		t.Fatal(err)
	}

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	q := u.Query()
	q.Set("id", id.String())
	q.Set("title", t.Name())
	q.Set("description", "a description")
	u.RawQuery = q.Encode()

	status, result := handlerTest(handleTaxiiCollections, "POST", u.String())
	if status != 200 {
		t.Fatal(result)
	}

	// test reading it
	u, err = url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	status, result = handlerTest(handleTaxiiCollections, "GET", u.String())

	expected := `{"id":"` + id.String() + `","can_read":true,"can_write":true,` +
		`"title":"` + t.Name() + `","description":"a description",` +
		`"media_types":["` + taxiiContentType + `"]}`

	if status != 200 {
		t.Error("Got:", status, "Expected:", 200)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetCollections(t *testing.T) {
	setupSQLite()

	u, err := url.Parse("https://localhost/api_root/collections/")
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections, "GET", u.String())

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
	u, err := url.Parse("https://localhost/api_root/collections/fail")
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections, "GET", u.String())
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

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	status, result := handlerTest(handleTaxiiCollections, "GET", u.String())
	expected := `{"title":"Resource not found","description":"Invalid Collection","http_status":"404"}` + "\n"

	if status != 404 {
		t.Error("Got:", status, "Expected:", 404)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetInvalidUser(t *testing.T) {
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
	handleTaxiiCollections(res, req)
	b, _ := ioutil.ReadAll(res.Body)

	status, result := res.Code, string(b)
	expected := `{"title":"Bad Request","description":"Invalid user specified","http_status":"400"}` + "\n"

	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestHandleTaxiiCollectionsGetReadError(t *testing.T) {
	defer setupSQLite()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	u, err := url.Parse("https://localhost/api_root/collections/" + id.String())
	if err != nil {
		t.Fatal(err)
	}

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections, "GET", u.String())
	if status != 400 {
		t.Error("Got:", status, "Expected:", 400)
	}
}

func TestHandleTaxiiCollectionsGetNoResults(t *testing.T) {
	defer setupSQLite()

	u, err := url.Parse("https://localhost/api_root/collections/")
	if err != nil {
		t.Fatal(err)
	}

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("delete from taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	status, _ := handlerTest(handleTaxiiCollections, "GET", u.String())
	if status != 404 {
		t.Error("Got:", status, "Expected:", 404)
	}
}

func TestReadTaxiiCollectionsFailRead(t *testing.T) {
	defer setupSQLite()

	w := httptest.NewRecorder()
	user := "foo"

	s, err := newSQLiteDB()
	if err != nil {
		t.Error(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_user_collection")
	if err != nil {
		t.Fatal(err)
	}

	readTaxiiCollections(w, user)

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

	cid, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: cid, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(testUser, "api_root")
	if err != nil {
		t.Fatal(err)
	}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
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
	err = s.db.QueryRow("select id from taxii_collection_api_root where id = '" + cid.String() + "'").Scan(&uid)
	if err != nil {
		t.Fatal(err)
	}

	if uid != cid.String() {
		t.Error("Got:", uid, "Expected:", cid.String())
	}
}

func TestTaxiiCollectionCreateFailTaxiiStorer(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}

	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}
	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiCollectionCreateFailQuery(t *testing.T) {
	renameFile("backend/sqlite/create/taxiiCollection.sql", "backend/sqlite/create/taxiiCollection.sql.testing")
	defer renameFile("backend/sqlite/create/taxiiCollection.sql.testing", "backend/sqlite/create/taxiiCollection.sql")

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionCreateFailWrite(t *testing.T) {
	defer setupSQLite()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionCreateFailWritePart(t *testing.T) {
	defer setupSQLite()

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	testCollection := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection_api_root")
	if err != nil {
		t.Fatal(err)
	}

	err = testCollection.create(testUser, "api_root")
	if err == nil {
		t.Error("Expected a write error")
	}
}

func TestTaxiiCollectionRead(t *testing.T) {
	setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
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
	err = tc.read(testUser)

	if tc.Title != "a title" {
		t.Error("Got:", tc.Title, "Expected", "a title")
	}
}

func TestTaxiiCollectionReadFail(t *testing.T) {
	defer reloadTestConfig()
	config = cabbyConfig{}
	config.DataStore = map[string]string{"name": "sqlite"}

	id, err := newTaxiiID()
	if err != nil {
		t.Fatal(err)
	}
	tc := taxiiCollection{ID: id, Title: "test collection", Description: "a test collection"}

	err = tc.read("user")
	if err == nil {
		t.Error("Expected error")
	}

	// fail due to no query being avialable
	defer renameFile("backend/sqlite/read/taxiiCollection.sql.testing", "backend/sqlite/read/taxiiCollection.sql")
	renameFile("backend/sqlite/read/taxiiCollection.sql", "backend/sqlite/read/taxiiCollection.sql.testing")

	reloadTestConfig()
	err = tc.read("user")
	if err == nil {
		t.Error("Expected error")
	}
}

func TestTaxiiCollectionsRead(t *testing.T) {
	setupSQLite()

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
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
	err = tcs.read(testUser)
	if err != nil {
		t.Fatal(err)
	}

	if len(tcs.Collections) != 1 {
		t.Error("Got:", len(tcs.Collections), "Expected", 1)
	}
}

func TestTaxiiCollectionsReadFailTaxiiStorer(t *testing.T) {
	config = cabbyConfig{}
	defer reloadTestConfig()

	tcs := taxiiCollections{}
	err := tcs.read(testUser)
	if err == nil {
		t.Error("Expected a taxiiStorer error")
	}
}

func TestTaxiiCollectionsReadFailQuery(t *testing.T) {
	renameFile("backend/sqlite/read/taxiiCollections.sql", "backend/sqlite/read/taxiiCollections.sql.testing")
	defer renameFile("backend/sqlite/read/taxiiCollections.sql.testing", "backend/sqlite/read/taxiiCollections.sql")

	tcs := taxiiCollections{}
	err := tcs.read(testUser)
	if err == nil {
		t.Error("Expected a query error")
	}
}

func TestTaxiiCollectionsReadFailWrite(t *testing.T) {
	defer setupSQLite()

	tcs := taxiiCollections{}

	s, err := newSQLiteDB()
	if err != nil {
		t.Fatal(err)
	}
	defer s.disconnect()

	_, err = s.db.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	err = tcs.read(testUser)
	if err == nil {
		t.Error("Expected a write error")
	}
}
