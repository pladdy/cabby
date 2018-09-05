package sqlite

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

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

	results, err := s.Object(test.CollectionID.String(), string(test.ID), cabby.Filter{})
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareObject(results[0], test)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectServiceObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	expected := tester.Object

	results, err := s.Object(expected.CollectionID.String(), string(expected.ID), cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareObject(results[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestObjectServiceObjectFilter(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create an object with multiple versions
	id, _ := cabby.NewID()
	objectID := "malware--" + id.String()

	for i := 0; i < 5; i++ {
		t := time.Now().UTC()
		createObjectVersion(ds, objectID, t.Format(time.RFC3339Nano))
		time.Sleep(100 * time.Millisecond)
	}

	tests := []struct {
		filter          cabby.Filter
		expectedObjects int
	}{
		{cabby.Filter{IDs: objectID}, 5},
		{cabby.Filter{IDs: objectID, Versions: "all"}, 5},
		{cabby.Filter{IDs: objectID, Versions: "first"}, 1},
		{cabby.Filter{IDs: objectID, Versions: "last"}, 1},
		{cabby.Filter{}, 5},
		{cabby.Filter{Versions: "all"}, 5},
		{cabby.Filter{Versions: "first"}, 1},
		{cabby.Filter{Versions: "last"}, 1},
	}

	// use for collectionID
	expected := tester.Object

	for _, test := range tests {
		results, err := s.Object(expected.CollectionID.String(), objectID, test.filter)
		if err != nil {
			t.Error("Got:", err, "Expected no error", "Filter:", test.filter)
		}

		if len(results) != test.expectedObjects {
			t.Error("Got:", len(results), "Expected:", test.expectedObjects, "Filter:", test.filter)
		}
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

	_, err = s.Object(expected.CollectionID.String(), string(expected.ID), cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectsServiceObjectsFilter(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create more objects to ensure not paged by default
	ids := []cabby.ID{}
	for i := 0; i < 10; i++ {
		id, _ := cabby.NewID()
		ids = append(ids, id)
		createObject(ds, id.String())
	}

	tests := []struct {
		filter          cabby.Filter
		expectedObjects int
	}{
		{cabby.Filter{}, 11},
		{cabby.Filter{Types: "foo"}, 0},
		{cabby.Filter{Types: "foo,malware"}, 11},
		{cabby.Filter{IDs: ids[0].String()}, 1},
		{cabby.Filter{IDs: strings.Join([]string{ids[0].String(), ids[4].String(), ids[8].String()}, ",")}, 3},
	}

	for _, test := range tests {
		results, err := s.Objects(tester.Object.CollectionID.String(), &cabby.Range{First: -1, Last: -1}, test.filter)
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results) != test.expectedObjects {
			t.Error("Got:", len(results), "Expected:", test.expectedObjects, "Filter:", test.filter)
		}
	}
}

func TestObjectsServiceObjectsRange(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create more objects to ensure not paged by default
	for i := 0; i < 10; i++ {
		id, _ := cabby.NewID()
		createObject(ds, id.String())
	}

	totalObjects := 11

	tests := []struct {
		cabbyRange      cabby.Range
		expectedObjects int
	}{
		// setupSQLite() creates 1 object, 10 created above (11 total)
		{cabby.Range{First: -1, Last: -1}, 11},
		{cabby.Range{First: 0, Last: 5}, 6},
	}

	for _, test := range tests {
		results, err := s.Objects(tester.Object.CollectionID.String(), &test.cabbyRange, cabby.Filter{})
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results) != test.expectedObjects {
			t.Error("Got:", len(results), "Expected:", test.expectedObjects)
		}

		if int(test.cabbyRange.Total) != totalObjects {
			t.Error("Got:", test.cabbyRange.Total, "Expected:", totalObjects)
		}
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

	_, err = s.Objects(expected.CollectionID.String(), &cabby.Range{}, cabby.Filter{})
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

	_, err = s.Objects("fail", &cabby.Range{}, cabby.Filter{})
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

		results, _ := osv.Object(tester.CollectionID, string(expected.ID), cabby.Filter{})

		passed := tester.CompareObject(results[0], expected)
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

	// check objects were saved; use an invalid range to get all
	result, _ := osv.Objects(tester.CollectionID, &cabby.Range{First: -1, Last: -1}, cabby.Filter{})
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
