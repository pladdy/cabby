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

func TestRangeInjectorWithPagination(t *testing.T) {
	result := WithPagination("select 1")
	expected := "select 1\nlimit ? offset ?"

	if result != expected {
		t.Error("Got:", result, "Expected:", expected)
	}
}

func TestSQLiteBatchWriteSmall(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description, media_types)
					values (?, ?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)
	toWrite <- []interface{}{"test", "test api root", "test collection", "this is a test collection", "media type"}
	close(toWrite)

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

func TestSQLiteBatchWriteExecuteLarge(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)

	recordsToWrite := 1000
	for i := 0; i <= recordsToWrite; i++ {
		toWrite <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError != nil {
		t.Error("Got:", lastError, "Expected: no error")
	}
}

func TestSQLiteBatchWriteWriteOperationError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchWrite("fail", toWrite, errs)
	toWrite <- []interface{}{"fail"}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchWriteExecuteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	sql := `insert into taxii_collection (id, api_root_path, title, description)
					values (?, ?, ?, ?)`

	go ds.batchWrite(sql, toWrite, errs)

	for i := 0; i <= maxWritesPerBatch; i++ {
		if i == maxWritesPerBatch {
			// a commit is about to happen
			ds.Close()
		}
		toWrite <- []interface{}{"test" + string(i), "api root", "collection", "a test collection"}
	}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestSQLiteBatchWriteCommitError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	toWrite := make(chan interface{}, 10)
	errs := make(chan error, 10)

	go ds.batchWrite("insert into stix_objects (id, object) values (?, ?)", toWrite, errs)
	toWrite <- []interface{}{"fail"}
	close(toWrite)

	var lastError error
	for e := range errs {
		lastError = e
	}

	if lastError == nil {
		t.Error("Expected error")
	}
}

func TestDataStoreWriteError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()

	err := ds.write("this is not a valid query")
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
