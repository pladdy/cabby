package sqlite

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io/ioutil"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

func TestObjectServiceCreateObject(t *testing.T) {
	// see TestObjectServiceDeleteObject; have to create and verify before deleting
}

func TestObjectServiceCreateObjectFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	obj := tester.GenerateObject("malware")
	err = s.CreateObject(context.Background(), tester.CollectionID, obj)
	if err == nil {
		t.Error("Expected error")
	}
}

func TestObjectServiceDeleteObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create object
	testObject := tester.GenerateObject("malware")
	err := s.CreateObject(context.Background(), tester.CollectionID, testObject)
	if err != nil {
		t.Error("Got:", err)
	}

	// verify object was created
	results, err := s.Object(context.Background(), tester.CollectionID, testObject.ID.String(), cabby.Filter{})
	if err != nil {
		t.Error("Got:", err)
	}

	passed := tester.CompareObject(results[0], testObject)
	if !passed {
		t.Error("Comparison failed")
	}

	// delete object
	err = s.DeleteObject(context.Background(), tester.CollectionID, testObject.ID.String())
	if err != nil {
		t.Error("Got:", err)
	}

	// verify object is gone
	results, err = s.Object(context.Background(), tester.CollectionID, testObject.ID.String(), cabby.Filter{})
	if err != nil {
		t.Error("Got:", err)
	}

	if len(results) > 0 {
		t.Error("Object", testObject.ID.String(), "was not deleted from collection", tester.CollectionID)
	}
}

func TestObjectServiceDeleteObjectFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	obj := tester.GenerateObject("malware")
	err = s.DeleteObject(context.Background(), tester.CollectionID, obj.ID.String())
	if err == nil {
		t.Error("Expected error")
	}
}

func TestObjectServiceObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	expected := tester.Object

	results, err := s.Object(context.Background(), tester.CollectionID, expected.ID.String(), cabby.Filter{})
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
	versions := []string{}

	for i := 0; i < 5; i++ {
		t := time.Now().UTC()
		createObjectVersion(ds, objectID, t.Format(time.RFC3339Nano))
		versions = append(versions, t.Format(time.RFC3339Nano))
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
		{cabby.Filter{Versions: versions[2]}, 1},
	}

	for _, test := range tests {
		results, err := s.Object(context.Background(), tester.CollectionID, objectID, test.filter)
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

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	expected := tester.Object

	_, err = s.Object(context.Background(), tester.CollectionID, expected.ID.String(), cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectsServiceObjectsFilter(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create more objects to ensure not paged by default
	ids := []stones.Identifier{}
	for i := 0; i < 10; i++ {
		id, _ := stones.NewIdentifier("malware")
		ids = append(ids, id)
		createObject(ds, id.String())
	}

	// create an object with multiple versions
	id, _ := stones.NewIdentifier("malware")
	objectID := "malware--" + id.String()
	versions := []string{}

	for i := 0; i < 5; i++ {
		t := time.Now().UTC()
		createObjectVersion(ds, objectID, t.Format(time.RFC3339Nano))
		versions = append(versions, t.Format(time.RFC3339Nano))
		time.Sleep(100 * time.Millisecond)
	}

	ts, err := stones.TimestampFromString(versions[0])
	if err != nil {
		t.Fatal(err)
	}

	tests := []struct {
		filter          cabby.Filter
		expectedObjects int
	}{
		{cabby.Filter{}, 16},
		{cabby.Filter{Types: "foo"}, 0},
		{cabby.Filter{Types: "foo,malware"}, 16},
		{cabby.Filter{IDs: ids[0].String()}, 1},
		{cabby.Filter{IDs: strings.Join([]string{ids[0].String(), ids[4].String(), ids[8].String()}, ",")}, 3},
		{cabby.Filter{Versions: versions[2]}, 1},
		{cabby.Filter{AddedAfter: ts}, 5},
	}

	cr := cabby.Page{}
	expectedTime := time.Now()

	for _, test := range tests {
		results, err := s.Objects(
			context.Background(),
			tester.CollectionID,
			&cabby.Page{}, test.filter)
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results) != test.expectedObjects {
			t.Error("Got:", len(results), "Expected:", test.expectedObjects, "Filter:", test.filter)
		}
	}

	if cr.MinimumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", cr.MinimumAddedAfter, "Expected >=:", expectedTime.UnixNano())
	}

	if cr.MaximumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", cr.MaximumAddedAfter, "Expected NOT:", expectedTime.UnixNano())
	}
}

func TestObjectsServiceObjectsPage(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// create more objects to ensure not paged by default
	for i := 0; i < 10; i++ {
		id, _ := stones.NewIdentifier("malware")
		createObject(ds, id.String())
	}

	totalObjects := 11

	tests := []struct {
		cabbyPage       cabby.Page
		expectedObjects int
	}{
		// setupSQLite() creates 1 object, 10 created above (11 total)
		{cabby.Page{Limit: 11}, 11},
		{cabby.Page{Limit: 6}, 6},
	}

	for _, test := range tests {
		results, err := s.Objects(context.Background(), tester.CollectionID, &test.cabbyPage, cabby.Filter{})
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results) != test.expectedObjects {
			t.Error("Got:", len(results), "Expected:", test.expectedObjects)
		}

		if int(test.cabbyPage.Total) != totalObjects {
			t.Error("Got:", test.cabbyPage.Total, "Expected:", totalObjects)
		}
	}
}

func TestObjectsServiceObjectsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Objects(context.Background(), tester.CollectionID, &cabby.Page{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestObjectServiceCreateEnvelope(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	osv := ds.ObjectService()
	ssv := ds.StatusService()

	envelopeFile, _ := os.Open("testdata/malware_envelope.json")
	content, _ := ioutil.ReadAll(envelopeFile)

	var envelope cabby.Envelope
	err := json.Unmarshal(content, &envelope)
	if err != nil {
		t.Fatal(err)
	}

	st := tester.Status
	osv.CreateEnvelope(context.Background(), envelope, tester.CollectionID, st, ssv)

	// check objects were saved
	for _, raw := range envelope.Objects {
		var expected stones.Object
		err := json.Unmarshal(raw, &expected)
		if err != nil {
			t.Fatal(err)
		}

		results, err := osv.Object(context.Background(), tester.CollectionID, expected.ID.String(), cabby.Filter{})
		if err != nil {
			t.Fatal(err)
		}

		passed := tester.CompareObject(results[0], expected)
		if !passed {
			t.Error("Comparison failed")
		}
	}
}

func TestObjectServiceCreateEnvelopeWithInvalidObject(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	osv := ds.ObjectService()
	ssv := ds.StatusService()

	// clear out any objects
	_, err := ds.DB.Exec("delete from objects")
	if err != nil {
		t.Fatal(err)
	}

	envelopeFile, _ := os.Open("testdata/invalid_objects_envelope.json")
	content, _ := ioutil.ReadAll(envelopeFile)

	var envelope cabby.Envelope
	err = json.Unmarshal(content, &envelope)
	if err != nil {
		t.Fatal(err)
	}

	st := tester.Status
	osv.CreateEnvelope(context.Background(), envelope, tester.CollectionID, st, ssv)

	// check objects were saved; use an invalid range to get all
	result, _ := osv.Objects(context.Background(), tester.CollectionID, &cabby.Page{}, cabby.Filter{})
	expected := 2
	if len(result) != expected {
		t.Error("Got:", len(result), "Expected:", expected)
	}
}

func TestObjectServiceInvalidIDs(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ObjectService()

	// change the underlying view the service depends on to return bad ids
	_, err := ds.DB.Exec(`drop view objects_data`)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ds.DB.Exec(
		`create view objects_data as select
		   1 rowid,
			 'fail' id,
			 'fail' type,
			 'fail' created,
			 'fail' modified,
			 'fail' collection_id,
			 'fail' object,
			 'fail' created_at,
			 1 count`,
	)
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Objects(context.Background(), "fail", &cabby.Page{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}

	_, err = s.Object(context.Background(), "fail", "fail", cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestUnmarshalObjectFail(t *testing.T) {
	tests := []struct {
		inputs      []string
		expectError bool
	}{
		{[]string{"malware--82407036-edf9-4c75-9a56-e72697c53e99", "invalid time", "2016-04-06T20:03:48.000Z"}, true},
		{[]string{"malware--82407036-edf9-4c75-9a56-e72697c53e99", "2016-04-06T20:03:48.000Z", "invalid time"}, true},
		{[]string{"invalid id", "2016-04-06T20:03:48.000Z", "2016-04-06T20:03:48.000Z"}, true},
		{[]string{"malware--82407036-edf9-4c75-9a56-e72697c53e99", "2016-04-06T20:03:48.000Z", "2016-04-06T20:03:48.000Z"}, false},
	}

	for _, test := range tests {
		_, err := unmarshalObject(stones.Object{}, test.inputs[0], test.inputs[1], test.inputs[2])

		if test.expectError && err == nil {
			t.Error("Expected an error", "Test:", test)
		}

		if !test.expectError && err != nil {
			t.Error("Error unexpected:", err, "Test:", test)
		}
	}
}

func TestUpdateStatusNoError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	ss := ds.StatusService()

	// assumptions:
	//   - a user posted an envelope of 3 objects
	expected := tester.Status
	expected.ID, _ = cabby.NewID()
	expected.Status = "pending"
	expected.TotalCount = 3

	err := ss.CreateStatus(context.Background(), expected)
	if err != nil {
		t.Fatal(err)
	}

	// check the pending status
	result, _ := ss.Status(context.Background(), expected.ID.String())
	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// no errors
	errs := make(chan error, 10)
	close(errs)

	// updating implies complete
	updateStatus(context.Background(), expected, errs, ss)

	expected.FailureCount = 0
	expected.PendingCount = 0
	expected.SuccessCount = 3
	expected.Status = "complete"

	// query the status to confirm it's accurate
	result, _ = ss.Status(context.Background(), expected.ID.String())
	passed = tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUpdateStatusWithError(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	ss := ds.StatusService()

	// assumptions:
	//   - a user posted an envelope of 3 objects
	expected := tester.Status
	expected.ID, _ = cabby.NewID()
	expected.Status = "pending"
	expected.TotalCount = 3

	err := ss.CreateStatus(context.Background(), expected)
	if err != nil {
		t.Fatal(err)
	}

	// check the pending status
	result, _ := ss.Status(context.Background(), expected.ID.String())
	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// assume one object failed to write
	errs := make(chan error, 10)
	errs <- errors.New("an error")
	close(errs)

	// updating implies complete
	updateStatus(context.Background(), expected, errs, ss)

	expected.FailureCount = 1
	expected.PendingCount = 0
	expected.SuccessCount = 2
	expected.Status = "complete"

	// query the status to confirm it's accurate
	result, _ = ss.Status(context.Background(), expected.ID.String())
	passed = tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUpdateStatusWithTooManyErrors(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	ss := ds.StatusService()

	// assumptions:
	//   - a user posted an envelope of 3 objects
	expected := tester.Status
	expected.ID, _ = cabby.NewID()
	expected.Status = "pending"
	expected.TotalCount = 1

	err := ss.CreateStatus(context.Background(), expected)
	if err != nil {
		t.Fatal(err)
	}

	// check the pending status
	result, _ := ss.Status(context.Background(), expected.ID.String())
	passed := tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}

	// create more errors than objects
	errs := make(chan error, 10)
	errs <- errors.New("an error")
	errs <- errors.New("an error")
	close(errs)

	// updating implies complete
	updateStatus(context.Background(), expected, errs, ss)

	expected.FailureCount = 1
	expected.PendingCount = 0
	expected.SuccessCount = 0
	expected.Status = "complete"

	// query the status to confirm it's accurate
	result, _ = ss.Status(context.Background(), expected.ID.String())
	passed = tester.CompareStatus(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestUpdateStatusFail(t *testing.T) {
	// redirect log output for test
	var buf bytes.Buffer

	log.SetFormatter(&log.JSONFormatter{})
	log.SetOutput(&buf)

	defer func() {
		log.SetFormatter(&log.TextFormatter{})
		log.SetOutput(os.Stderr)
	}()

	// set up db
	setupSQLite()
	ds := testDataStore()
	ss := ds.StatusService()

	err := ss.CreateStatus(context.Background(), tester.Status)
	if err != nil {
		t.Fatal(err)
	}

	_, err = ds.DB.Exec("drop table status")
	if err != nil {
		t.Fatal(err)
	}

	// assume one object failed to write
	errs := make(chan error, 10)
	errs <- errors.New("an error")
	close(errs)
	updateStatus(context.Background(), tester.Status, errs, ss)

	// parse log into struct
	var result tester.ErrorLog
	err = json.Unmarshal([]byte(tester.LastLog(buf)), &result)
	if err != nil {
		t.Fatal(err)
	}

	if result.Level != "error" {
		t.Error("Got:", result.Level, "Expected: error")
	}
}
