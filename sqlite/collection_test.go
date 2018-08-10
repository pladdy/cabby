package sqlite

import (
	"strings"
	"testing"
)

func TestCollectionServiceCollection(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	expected := testCollection()

	result, err := s.Collection(testUserEmail, testCollectionID)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if result.ID.String() != expected.ID.String() {
		t.Error("Got:", result.ID.String(), "Expected:", expected.ID.String())
	}
	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		t.Error("Got:", result.Description, "Expected:", expected.Description)
	}
	if result.CanRead != expected.CanRead {
		t.Error("Got:", result.CanRead, "Expected:", expected.CanRead)
	}
	if result.CanWrite != expected.CanWrite {
		t.Error("Got:", result.CanWrite, "Expected:", expected.CanWrite)
	}
	if strings.Join(result.MediaTypes, ",") != strings.Join(expected.MediaTypes, ",") {
		t.Error("Got:", strings.Join(result.MediaTypes, ","), "Expected:", strings.Join(expected.MediaTypes, ","))
	}
}

func TestCollectionServiceCollectionQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collection(testUserEmail, testCollectionID)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}

func TestCollectionsServiceCollections(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	expected := testCollection()

	result, err := s.Collections(testUserEmail)
	if err != nil {
		t.Error("Got:", err, "Expected no error")
	}

	if len(result.Collections) <= 0 {
		t.Error("Got:", len(result.Collections), "Expected: > 0 Collections")
	}

	c := result.Collections[0]

	if c.ID.String() != expected.ID.String() {
		t.Error("Got:", c.ID.String(), "Expected:", expected.ID.String())
	}
	if c.Title != expected.Title {
		t.Error("Got:", c.Title, "Expected:", expected.Title)
	}
	if c.Description != expected.Description {
		t.Error("Got:", c.Description, "Expected:", expected.Description)
	}
	if c.CanRead != expected.CanRead {
		t.Error("Got:", c.CanRead, "Expected:", expected.CanRead)
	}
	if c.CanWrite != expected.CanWrite {
		t.Error("Got:", c.CanWrite, "Expected:", expected.CanWrite)
	}
	if strings.Join(c.MediaTypes, ",") != strings.Join(expected.MediaTypes, ",") {
		t.Error("Got:", strings.Join(c.MediaTypes, ","), "Expected:", strings.Join(expected.MediaTypes, ","))
	}
}

func TestCollectionsServiceCollectionsQueryErr(t *testing.T) {
	setupSQLite()
	ds := testDataStore()
	s := CollectionService{DB: ds.DB}

	_, err := s.DB.Exec("drop table taxii_collection")
	if err != nil {
		t.Fatal(err)
	}

	_, err = s.Collections(testUserEmail)
	if err == nil {
		t.Error("Got:", err, "Expected an error")
	}
}
