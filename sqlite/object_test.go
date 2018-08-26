package sqlite

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
	"github.com/pladdy/stones"
)

func TestBytesToObjectValidJSON(t *testing.T) {
	_, err := bytesToObject([]byte(`{
		"type": "malware",
		"id": "malware--31b940d4-6f7f-459a-80ea-9c1f17b5891b",
    "created": "2016-04-06T20:07:09.000Z",
		"modified": "2016-04-06T20:07:09.000Z",
		"name": "Poison Ivy"}`))

	if err != nil {
		t.Error("Expected no error")
	}
}

func TestBytesToObjectInvalidObject(t *testing.T) {
	_, err := bytesToObject([]byte(`{"foo": "bar"}`))
	if err == nil {
		t.Error("Expected an error")
	}
}

func TestBytesToObjectInvalidJSON(t *testing.T) {
	o, err := bytesToObject([]byte(`{"foo": "bar"`))
	if err == nil {
		t.Error("Expected error for bundle:", o)
	}
}

func TestObjectServiceCreateObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	test := tester.GenerateObject("malware")

	err := s.CreateObject(test)
	if err != nil {
		t.Error("Got:", err)
	}

	result, err := s.Object(test.CollectionID.String(), string(test.ID))
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareObject(result, test)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectServiceObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	expected := tester.Object

	result, err := s.Object(expected.CollectionID.String(), string(expected.ID))
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareObject(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectServiceObjectQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Object

	_, err = s.Object(expected.CollectionID.String(), string(expected.ID))
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectsServiceObjects(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	expected := tester.Object

	results, err := s.Objects(expected.CollectionID.String())
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(results) <= 0 {
		t.Error("Got:", len(results), "Expected: > 0")
	}

	passed := tester.CompareObject(results[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectsServiceObjectsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Object

	_, err = s.Objects(expected.CollectionID.String())
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectServiceObjectsInvalidID(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec(`insert into stix_objects (id, type, created, modified, object, collection_id)
	                     values ('fail', 'fail', 'fail', 'fail', '{"fail": true}', 'fail')`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Objects("fail")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectServiceCreateBundle(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	osv := ds.ObjectService()
	ssv := ds.StatusService()

	bundleFile, _ := os.Open("testdata/malware_bundle.json")
	content, _ := ioutil.ReadAll(bundleFile)

	var bundle stones.Bundle
	err := json.Unmarshal(content, &bundle)
	if err != nil {
		t.Fatal(err)
	}

	st := tester.Status
	osv.CreateBundle(bundle, tester.CollectionID, st, ssv)

	// check objects were saved
	for _, object := range bundle.Objects {
		expected, _ := bytesToObject(object)
		expected.CollectionID = tester.Collection.ID

		result, _ := osv.Object(tester.CollectionID, string(expected.ID))

		passed := tester.CompareObject(result, expected)
		if !passed {
			t.Error("Comparison failed")
		}
	}
}

func TestObjectServiceCreateBundleWithInvalidObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	osv := ds.ObjectService()
	ssv := ds.StatusService()

	// clear out any objects
	_, err := ds.DB.Exec("delete from stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	bundleFile, _ := os.Open("testdata/bundle_invalid_object.json")
	content, _ := ioutil.ReadAll(bundleFile)

	var bundle stones.Bundle
	err = json.Unmarshal(content, &bundle)
	if err != nil {
		t.Fatal(err)
	}

	st := tester.Status
	osv.CreateBundle(bundle, tester.CollectionID, st, ssv)

	// check objects were saved
	result, _ := osv.Objects(tester.CollectionID)
	expected := 2
	if len(result) != expected {
		t.Error("Got:", len(result), "Expected:", expected)
	}
}

func TestUpdateStatus(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	ss := ds.StatusService()

	// assumptions:
	//   - a status is already created for a user
	//   - a user posted a bundle of 3 objects
	expected := tester.Status
	expected.ID, _ = cabby.NewID()
	expected.Status = "pending"
	expected.TotalCount = 3

	err := ss.CreateStatus(expected)
	if err != nil {
		t.Fatal(err)
	}

	// assume one object failed to write
	errs := make(chan error, 10)
	errs <- errors.New("an error")
	close(errs)
	updateStatus(expected, errs, ss)

	expected.PendingCount = 2
	expected.FailureCount = 1

	// query the status to confirm it's accurate
	result, _ := ss.Status(expected.ID.String())
	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}
