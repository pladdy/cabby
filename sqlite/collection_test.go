package sqlite

import (
	"testing"

	"github.com/pladdy/cabby2/tester"
)

func TestCollectionServiceCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	expected := tester.Collection

	result, err := s.Collection(tester.UserEmail, expected.APIRootPath, expected.ID.String())
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	tester.CompareCollection(result, expected, t)
}

func TestCollectionServiceCollectionQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_collection")
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
	s := CollectionService{DB: ds.DB}

	expected := tester.Collection

	results, err := s.Collections(tester.UserEmail, tester.APIRootPath)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(results.Collections) <= 0 {
		t.Error("Got:", len(results.Collections), "Expected: > 0 Collections")
	}

	result := results.Collections[0]

	tester.CompareCollection(result, expected, t)
}

func TestCollectionsServiceCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collections(tester.UserEmail, tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
