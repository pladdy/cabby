package sqlite

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
	"github.com/pladdy/stones"
)

func TestManifestServiceManifest(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	cr := cabby.Range{}
	expected := tester.ManifestEntry
	expectedTime := time.Now().UTC()

	result, err := s.Manifest(context.Background(), tester.CollectionID, &cr, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	// sync the date added times; this is set in the db, not sure how to "mock" this
	expected.DateAdded = result.Objects[0].DateAdded

	passed := tester.CompareManifestEntry(result.Objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}

	if cr.MinimumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", cr.MinimumAddedAfter.UnixNano(), "Expected >=:", expectedTime.UnixNano())
	}

	if cr.MaximumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", cr.MaximumAddedAfter.UnixNano(), "Expected NOT:", expectedTime.UnixNano())
	}
}

func TestManifestServiceManifestFilter(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	// create more objects to ensure not paged by default
	ids := []stones.Identifier{}
	for i := 0; i < 10; i++ {
		id, _ := stones.NewIdentifier("malware")
		ids = append(ids, id)
		createObject(ds, id.String())
	}

	tests := []struct {
		filter          cabby.Filter
		expectedEntries int
	}{
		{cabby.Filter{}, 11},
		{cabby.Filter{Types: "foo"}, 0},
		{cabby.Filter{Types: "foo,malware"}, 11},
		{cabby.Filter{IDs: ids[0].String()}, 1},
		{cabby.Filter{IDs: strings.Join([]string{ids[0].String(), ids[4].String(), ids[8].String()}, ",")}, 3},
	}

	for _, test := range tests {
		results, err := s.Manifest(context.Background(), tester.Collection.ID.String(), &cabby.Range{First: 0, Last: 0}, test.filter)
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results.Objects) != test.expectedEntries {
			t.Error("Got:", len(results.Objects), "Expected:", test.expectedEntries)
		}
	}
}

func TestManifestServiceManifestRange(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	// create more objects to ensure not paged by default
	for i := 0; i < 10; i++ {
		id, _ := stones.NewIdentifier("malware")
		createObject(ds, id.String())
	}

	totalEntries := 11

	tests := []struct {
		cabbyRange      cabby.Range
		expectedEntries int
	}{
		// setupSQLite() creates 1 object, 10 created above (11 total)
		{cabby.Range{First: 0, Last: 0}, 11},
		{cabby.Range{First: 0, Last: 5, Set: true}, 6},
	}

	for _, test := range tests {
		results, err := s.Manifest(context.Background(), tester.Collection.ID.String(), &test.cabbyRange, cabby.Filter{})
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results.Objects) != test.expectedEntries {
			t.Error("Got:", len(results.Objects), "Expected:", test.expectedEntries)
		}

		if int(test.cabbyRange.Total) != totalEntries {
			t.Error("Got:", test.cabbyRange.Total, "Expected:", totalEntries)
		}
	}
}

func TestManifestServiceManifestQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	_, err := ds.DB.Exec("drop table objects")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Manifest(context.Background(), tester.CollectionID, &cabby.Range{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestManifestServiceManifestNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	_, err := s.Manifest(context.Background(), tester.CollectionID, &cabby.Range{}, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
