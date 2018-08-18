package tester

import (
	"strings"

	cabby "github.com/pladdy/cabby2"
)

// CompareAPIRoot compares two APIRoots
func CompareAPIRoot(result, expected cabby.APIRoot) bool {
	passed := true

	if result.Path != expected.Path {
		Error.Println("Got:", result.Path, "Expected:", expected.Path)
	}
	if result.Title != expected.Title {
		Error.Println("Got:", result.Title, "Expected:", expected.Title)
	}
	if result.Description != expected.Description {
		Error.Println("Got:", result.Description, "Expected:", expected.Description)
	}
	if strings.Join(result.Versions, ",") != strings.Join(expected.Versions, ",") {
		Error.Println("Got:", strings.Join(result.Versions, ","), "Expected:", strings.Join(expected.Versions, ","))
	}
	if result.MaxContentLength != expected.MaxContentLength {
		Error.Println("Got:", result.MaxContentLength, "Expected:", expected.MaxContentLength)
	}

	return passed
}

// CompareCollection compares two Collections
func CompareCollection(result, expected cabby.Collection) bool {
	passed := true

	if result.ID.String() != expected.ID.String() {
		Error.Println("Got:", result.ID.String(), "Expected:", expected.ID.String())
		passed = false
	}
	if result.Title != expected.Title {
		Error.Println("Got:", result.Title, "Expected:", expected.Title)
		passed = false
	}
	if result.Description != expected.Description {
		Error.Println("Got:", result.Description, "Expected:", expected.Description)
		passed = false
	}
	if result.CanRead != expected.CanRead {
		Error.Println("Got:", result.CanRead, "Expected:", expected.CanRead)
		passed = false
	}
	if result.CanWrite != expected.CanWrite {
		Error.Println("Got:", result.CanWrite, "Expected:", expected.CanWrite)
		passed = false
	}
	if strings.Join(result.MediaTypes, ",") != strings.Join(expected.MediaTypes, ",") {
		Error.Println("Got:", strings.Join(result.MediaTypes, ","), "Expected:", strings.Join(expected.MediaTypes, ","))
		passed = false
	}

	return passed
}

// CompareDiscovery compares two Discoverys
func CompareDiscovery(result, expected cabby.Discovery) bool {
	passed := true

	if result.Title != expected.Title {
		Error.Println("Got:", result.Title, "Expected:", expected.Title)
		passed = false
	}
	if result.Description != expected.Description {
		Error.Println("Got:", result.Description, "Expected:", expected.Description)
		passed = false
	}
	if result.Contact != expected.Contact {
		Error.Println("Got:", result.Contact, "Expected:", expected.Contact)
		passed = false
	}

	if result.Default != expected.Default {
		Error.Println("Got:", result.Default, "Expected:", expected.Default)
		passed = false
	}
	if result.APIRoots[0] != expected.APIRoots[0] {
		Error.Println("Got:", result.APIRoots[0], "Expected:", expected.APIRoots[0])
		passed = false
	}

	return passed
}

// CompareError compares two Errors
func CompareError(result, expected cabby.Error) bool {
	passed := true

	if result.Title != expected.Title {
		Error.Println("Got:", result.Title, "Expected:", expected.Title)
		passed = false
	}
	if result.Description != expected.Description {
		Error.Println("Got:", result.Description, "Expected:", expected.Description)
		passed = false
	}
	if result.HTTPStatus != expected.HTTPStatus {
		Error.Println("Got:", result.HTTPStatus, "Expected:", expected.HTTPStatus)
		passed = false
	}

	return passed
}

// CompareObject compares two Collections
func CompareObject(result, expected cabby.Object) bool {
	passed := true

	if result.ID != expected.ID {
		Error.Println("Got:", result.ID, "Expected:", expected.ID)
		passed = false
	}
	if result.Type != expected.Type {
		Error.Println("Got:", result.Type, "Expected:", expected.Type)
		passed = false
	}
	if result.Created != expected.Created {
		Error.Println("Got:", result.Created, "Expected:", expected.Created)
		passed = false
	}
	if result.Modified != expected.Modified {
		Error.Println("Got:", result.Modified, "Expected:", expected.Modified)
		passed = false
	}

	rObject := string(result.Object)
	eObject := string(expected.Object)
	if rObject != eObject {
		Error.Println("Got:", rObject, "Expected:", eObject)
		passed = false
	}

	if result.CollectionID.String() != expected.CollectionID.String() {
		Error.Println("Got:", result.CollectionID.String(), "Expected:", expected.CollectionID.String())
		passed = false
	}

	return passed
}
