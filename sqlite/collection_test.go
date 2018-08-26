package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestCollectionServiceCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	expected := tester.Collection

	result, err := s.Collection(tester.UserEmail, expected.APIRootPath, expected.ID.String())
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	passed := tester.CompareCollection(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionServiceCollectionQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collection(tester.UserEmail, tester.CollectionID, tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionsServiceCollections(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	expected := tester.Collection

	results, err := s.Collections(tester.UserEmail, tester.APIRootPath)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(results.Collections) <= 0 {
		t.Error("Got:", len(results.Collections), "Expected: > 0 Collections")
	}

	result := results.Collections[0]

	passed := tester.CompareCollection(result, expected)
	if !passed {
		t.Error("Comparison failed")
	}
}

func TestCollectionsServiceCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collections(tester.UserEmail, tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionsServiceCollectionsInAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	expected := tester.Collection

	results, err := s.CollectionsInAPIRoot(tester.APIRootPath)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(results.CollectionIDs) <= 0 {
		t.Error("Got:", len(results.CollectionIDs), "Expected: > 0 Collections")
	}
	if results.Path != tester.APIRootPath {
		t.Error("Got:", results.Path, "Expected:", tester.APIRootPath)
	}

	// if more ids are added for other tests, this loop has to be updated
	for _, id := range results.CollectionIDs {
		if id.String() != expected.ID.String() {
			t.Error("Got:", id.String(), "Expected:", expected.ID.String())
		}
	}
}

func TestCollectionsServiceCollectionsInAPIRootQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.CollectionsInAPIRoot(tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
