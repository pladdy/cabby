package sqlite

import (
	"strings"
	"testing"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/tester"
)

func TestManifestServiceManifest(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	expected := tester.ManifestEntry

	result, err := s.Manifest(tester.CollectionID, &cabby.Range{}, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareManifestEntry(result.Objects[0], expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestManifestServiceManifestFilter(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	// create more objects to ensure not paged by default
	ids := []cabby.ID{}
	for i := 0; i < 10; i++ {
		id, _ := cabby.NewID()
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
		results, err := s.Manifest(tester.Object.CollectionID.String(), &cabby.Range{First: -1, Last: -1}, test.filter)
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
		id, _ := cabby.NewID()
		createObject(ds, id.String())
	}

	totalEntries := 11

	tests := []struct {
		cabbyRange      cabby.Range
		expectedEntries int
	}{
		// setupSQLite() creates 1 object, 10 created above (11 total)
		{cabby.Range{First: -1, Last: -1}, 11},
		{cabby.Range{First: 0, Last: 5}, 6},
	}

	for _, test := range tests {
		results, err := s.Manifest(tester.Object.CollectionID.String(), &test.cabbyRange, cabby.Filter{})
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

	_, err := ds.DB.Exec("drop table stix_objects")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Manifest(tester.CollectionID, &cabby.Range{}, cabby.Filter{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestManifestServiceManifestNoAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.ManifestService()

	_, err := s.Manifest(tester.CollectionID, &cabby.Range{}, cabby.Filter{})
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}
}
