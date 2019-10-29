package sqlite

import (
	"context"
	"testing"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/tester"
)

func TestCollectionServiceCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	expected := tester.Collection

	result, err := s.Collection(tester.Context, expected.APIRootPath, expected.ID.String())
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

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collection(tester.Context, tester.CollectionID, tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionsServiceCollections(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	// create more collections to ensure not paged by default
	for i := 0; i < 10; i++ {
		id, _ := cabby.NewID()
		createCollection(ds, id.String())
	}

	totalCollections := 11

	tests := []struct {
		cabbyPage           cabby.Page
		expectedCollections int
	}{
		// setupSQLite() creates 1 collection, 10 created above (11 total)
		{cabby.Page{Limit: 11}, 11},
		{cabby.Page{Limit: 6}, 6},
	}

	for _, test := range tests {
		results, err := s.Collections(tester.Context, tester.APIRootPath, &test.cabbyPage)
		if err != nil {
			t.Error("Got:", err, "Expected no error")
		}

		if len(results.Collections) != test.expectedCollections {
			t.Error("Got:", len(results.Collections), "Expected:", test.expectedCollections)
		}

		if int(test.cabbyPage.Total) != totalCollections {
			t.Error("Got:", test.cabbyPage.Total, "Expected:", totalCollections)
		}
	}
}

func TestCollectionsServiceCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collections(tester.Context, tester.APIRootPath, &cabby.Page{})
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionsServiceCollectionsInAPIRoot(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	expected := tester.Collection

	results, err := s.CollectionsInAPIRoot(tester.Context, tester.APIRootPath)
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

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.CollectionsInAPIRoot(tester.Context, tester.APIRootPath)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionServiceCreateCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	newID, _ := cabby.NewID()
	expected := tester.Collection
	expected.ID = newID

	err := s.CreateCollection(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	rows, _ := ds.DB.Query("select id from collection where id = ?", expected.ID.String())
	defer rows.Close()

	var result string
	for rows.Next() {
		rows.Scan(&result)
	}

	if result != expected.ID.String() {
		t.Error("Got:", result, "Expected:", expected.ID.String())
	}
}

func TestUserServiceCreateUserInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	err := s.CreateCollection(context.Background(), cabby.Collection{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestCollectionServiceCreateCollectionQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	err = s.CreateCollection(context.Background(), tester.Collection)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionServiceDeleteCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	// create and verify a user
	collectionID, _ := cabby.NewID()
	expected := cabby.Collection{ID: collectionID, Title: "a title"}

	err := s.CreateCollection(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	rows, _ := ds.DB.Query("select id from collection where id = ?", expected.ID.String())
	defer rows.Close()

	var result string
	for rows.Next() {
		rows.Scan(&result)
	}

	if result != collectionID.String() {
		t.Error("collection not created")
	}

	// delete and verify collection is gone
	err = s.DeleteCollection(context.Background(), expected.ID.String())
	if err != nil {
		t.Error("Got:", err)
	}

	rows, _ = ds.DB.Query("select id from collection where id = ?", expected.ID.String())
	defer rows.Close()

	result = ""
	for rows.Next() {
		rows.Scan(&result)
	}

	if result != "" {
		t.Error("Got:", result, `Expected: ""`)
	}
}

func TestCollectionServiceDeleteCollectionQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	err = s.DeleteCollection(context.Background(), "foo")
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionServiceUpdateCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	newID, _ := cabby.NewID()
	expected := tester.Collection
	expected.ID = newID

	err := s.CreateCollection(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	expected.Title = "an updated title"
	expected.Description = "an updated description"

	err = s.UpdateCollection(context.Background(), expected)
	if err != nil {
		t.Error("Got:", err)
	}

	rows, _ := ds.DB.Query("select title, description from collection where id = ?", expected.ID.String())
	defer rows.Close()

	var title string
	var description string
	for rows.Next() {
		rows.Scan(&title, &description)
	}

	if title != expected.Title {
		t.Error("Got:", title, "Expected:", expected.Title)
	}
	if description != expected.Description {
		t.Error("Got:", title, "Expected:", expected.Description)
	}
}

func TestCollectionServiceUpdateCollectionInvalid(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	err := s.UpdateCollection(context.Background(), cabby.Collection{})
	if err == nil {
		t.Error("Expected an err")
	}
}

func TestCollectionServiceUpdateCollectionQueryFail(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := ds.CollectionService()

	_, err := ds.DB.Exec("drop table collection")
	if err != nil {
		t.Fatal(err)
	}

	err = s.UpdateCollection(context.Background(), tester.Collection)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
