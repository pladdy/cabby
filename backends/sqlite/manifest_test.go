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

	p := cabby.Page{}
	expected := tester.ManifestEntry
	expectedTime := time.Now().UTC()

	result, err := s.Manifest(context.Background(), tester.CollectionID, &p, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	// sync the date added times; this is set in the db, not sure how to "mock" this
	expected.DateAdded = result.Objects[0].DateAdded

	passed := tester.CompareManifestEntry(result.Objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}

	if p.MinimumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", p.MinimumAddedAfter.UnixNano(), "Expected >=:", expectedTime.UnixNano())
	}

	if p.MaximumAddedAfter.UnixNano() >= expectedTime.UnixNano() {
		t.Error("Got:", p.MaximumAddedAfter.UnixNano(), "Expected NOT:", expectedTime.UnixNano())
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
		results, err := s.Manifest(context.Background(), tester.Collection.ID.String(), &cabby.Page{}, test.filter)
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results.Objects) != test.expectedEntries {
			t.Error("Got:", len(results.Objects), "Expected:", test.expectedEntries)
		}
	}
}

func TestManifestServiceManifestPage(t *testing.T) {
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
		cabbyPage       cabby.Page
		expectedEntries int
	}{
		// setupSQLite() creates 1 object, 10 created above (11 total)
		{cabby.Page{Limit: 11}, 11},
		{cabby.Page{Limit: 6}, 6},
	}

	for _, test := range tests {
		results, err := s.Manifest(context.Background(), tester.Collection.ID.String(), &test.cabbyPage, cabby.Filter{})
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results.Objects) != test.expectedEntries {
			t.Error("Got:", len(results.Objects), "Expected:", test.expectedEntries)
		}

		if int(test.cabbyPage.Total) != totalEntries {
			t.Error("Got:", test.cabbyPage.Total, "Expected:", totalEntries)
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

	_, err = s.Manifest(context.Background(), tester.CollectionID, &cabby.Page{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestManifestServiceManifestNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	_, err := s.Manifest(context.Background(), tester.CollectionID, &cabby.Page{}, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}

func TestManifestServiceManifestInvalidDateAdded(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	_, err := ds.DB.Exec("update objects set created_at = 'fail'")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Manifest(context.Background(), tester.CollectionID, &cabby.Page{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
