package sqlite

import (
	"testing"
)

func TestNewDataStore(t *testing.T) {
	_, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}
}

func TestNewDataStoreNoPath(t *testing.T) {
	_, err := NewDataStore("")

	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreClose(t *testing.T) {
	s, err := NewDataStore("temp.db")

	if err != nil {
		t.Error("Got:", err, "Expected: nil")
	}

	s.Close()
}

func TestSQLiteBatchCreateSmall(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description, media_types)
					values (?, ?, ?, ?, ?)`

	go ds.batchCreate(sql, toCreate, errs)
	toCreate <- []interface{}{"test", "test api root", "test collection", "this is a test collection", "media type"}
	close(toCreate)

	for e := range errs {
		t.Fatal(e)
	}

	var id string
	err := ds.DB.QueryRow(`select id from taxii_collection where id = 'test'`).Scan(&id)
	if err != nil {
		t.Fatal(err)
	}

	if id != "test" {
		t.Error("Got:", id, "Expected:", "test")
	}
}

func TestSQLiteBatchCreateExecuteLarge(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchCreate(sql, toCreate, errs)

	recordsToCreate := 1000
	for i := 0; i <= recordsToCreate; i++ {
		toCreate <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toCreate)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError != nil {
		t.Error("Got:", lastError, "Expected: no error")
	}
}

func TestSQLiteBatchCreateWriteOperationError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchCreate("fail", toCreate, errs)
	toCreate <- []interface{}{"fail"}
	close(toCreate)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchCreateExecuteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchCreate(sql, toCreate, errs)

	for i := 0; i <= maxWritesPerBatch; i++ {
		if i == maxWritesPerBatch {
			// a commit is about to happen
			ds.Close()
		}
		toCreate <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toCreate)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchCreateCommitError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toCreate := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchCreate("insert into stix_objects (id, object) values (?, ?)", toCreate, errs)
	toCreate <- []interface{}{"fail"}
	close(toCreate)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestDataStoreCreateError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	err := ds.create("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreExecuteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	tx, _ := ds.DB.Begin()
	stmt, _ := tx.Prepare("select * from stix_objects")

	err := ds.execute(stmt, "fail")
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestDataStoreWriteOperationError(t *testing.T) {
	ds := testDataStore()

	ds.Open()
	_, _, err := ds.writeOperation("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}

	ds.Close()
	_, _, err = ds.writeOperation("this is not a valid query")
	if err == nil {
		t.Error("Expected an error")
	}
}
