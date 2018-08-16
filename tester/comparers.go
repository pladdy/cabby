package tester

import (
	"strings"
	"testing"

	cabby "github.com/pladdy/cabby2"
)

// CompareCollection compares two Collections
func CompareCollection(result, expected cabby.Collection, t *testing.T) {
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

// CompareDiscovery compares two Discoverys
func CompareDiscovery(result, expected cabby.Discovery, t *testing.T) {
	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		t.Error("Got:", result.Description, "Expected:", expected.Description)
	}
	if result.Contact != expected.Contact {
		t.Error("Got:", result.Contact, "Expected:", expected.Contact)
	}

	if result.Default != expected.Default {
		t.Error("Got:", result.Default, "Expected:", expected.Default)
	}
	if result.APIRoots[0] != expected.APIRoots[0] {
		t.Error("Got:", result.APIRoots[0], "Expected:", expected.APIRoots[0])
	}
}

// CompareError compares two Errors
func CompareError(result, expected cabby.Error, t *testing.T) {
	if result.Title != expected.Title {
		t.Error("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		t.Error("Got:", result.Description, "Expected:", expected.Description)
	}
	if result.HTTPStatus != expected.HTTPStatus {
		t.Error("Got:", result.HTTPStatus, "Expected:", expected.HTTPStatus)
	}
}

// CompareObject compares two Collections
func CompareObject(result, expected cabby.Object, t *testing.T) {
	if result.RawID != expected.RawID {
		t.Error("Got:", result.RawID, "Expected:", expected.RawID)
	}
	if result.ID.String() != expected.ID.String() {
		t.Error("Got:", result.ID.String(), "Expected:", expected.ID.String())
	}
	if result.Type != expected.Type {
		t.Error("Got:", result.Type, "Expected:", expected.Type)
	}
	if result.Created != expected.Created {
		t.Error("Got:", result.Created, "Expected:", expected.Created)
	}
	if result.Modified != expected.Modified {
		t.Error("Got:", result.Modified, "Expected:", expected.Modified)
	}

	rObject := string(result.Object)
	eObject := string(expected.Object)
	if rObject != eObject {
		t.Error("Got:", rObject, "Expected:", eObject)
	}

	if result.CollectionID.String() != expected.CollectionID.String() {
		t.Error("Got:", result.CollectionID.String(), "Expected:", expected.CollectionID.String())
	}
}
