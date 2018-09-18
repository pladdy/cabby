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

	for _, apiRoot := range result.APIRoots {
		if apiRoot != apiRoot {
			Error.Println("Got:", result.APIRoots[0], "Expected:", expected.APIRoots[0])
			passed = false
		}
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

// CompareManifestEntry compares two manifest entries
func CompareManifestEntry(result, expected cabby.ManifestEntry) bool {
	passed := true

	if result.ID != expected.ID {
		Error.Println("Got:", result.ID, "Expected:", expected.ID)
		passed = false
	}
	if result.DateAdded != expected.DateAdded {
		Error.Println("Got:", result.DateAdded, "Expected:", expected.DateAdded)
		passed = false
	}

	rVersions := strings.Join(result.Versions, ",")
	eVersions := strings.Join(expected.Versions, ",")
	if rVersions != eVersions {
		Error.Println("Got:", rVersions, "Expected:", eVersions)
		passed = false
	}

	rMediaTypes := strings.Join(result.MediaTypes, ",")
	eMediaTypes := strings.Join(expected.MediaTypes, ",")
	if rMediaTypes != eMediaTypes {
		Error.Println("Got:", rMediaTypes, "Expected:", eMediaTypes)
		passed = false
	}

	return passed
}

// CompareObject compares two objects
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

// CompareStatus compares two objects
func CompareStatus(result, expected cabby.Status) bool {
	passed := true

	if result.ID.String() != expected.ID.String() {
		Error.Println("Got:", result.ID.String(), "Expected:", expected.ID.String())
		passed = false
	}
	if result.Status != expected.Status {
		Error.Println("Got:", result.Status, "Expected:", expected.Status)
		passed = false
	}
	if result.RequestTimestamp != expected.RequestTimestamp {
		Error.Println("Got:", result.RequestTimestamp, "Expected:", expected.RequestTimestamp)
		passed = false
	}
	if result.TotalCount != expected.TotalCount {
		Error.Println("Got:", result.TotalCount, "Expected:", expected.TotalCount)
		passed = false
	}

	passed = compareFailures(result, expected, passed)
	passed = comparePendings(result, expected, passed)
	passed = compareSuccesses(result, expected, passed)
	return passed
}

// CompareUser compares two Users
func CompareUser(result, expected cabby.User) bool {
	passed := true

	if result.Email != expected.Email {
		Error.Println("Got:", result.Email, "Expected:", expected.Email)
		passed = false
	}
	if result.CanAdmin != expected.CanAdmin {
		Error.Println("Got:", result.CanAdmin, "Expected:", expected.CanAdmin)
		passed = false
	}

	return passed
}

/* compareStatus helpers */

func compareFailures(result, expected cabby.Status, passed bool) bool {
	if result.FailureCount != expected.FailureCount {
		Error.Println("Got:", result.FailureCount, "Expected:", expected.FailureCount)
		passed = false
	}

	for i := 0; i < len(result.Failures); i++ {
		if result.Failures[i] != expected.Failures[i] {
			Error.Println("Got:", result.Failures[i], "Expected:", expected.Failures[i])
			passed = false
		}
	}

	return passed
}

func comparePendings(result, expected cabby.Status, passed bool) bool {
	if result.PendingCount != expected.PendingCount {
		Error.Println("Got:", result.PendingCount, "Expected:", expected.PendingCount)
		passed = false
	}

	for i := 0; i < len(result.Pendings); i++ {
		if result.Pendings[i] != expected.Pendings[i] {
			Error.Println("Got:", result.Pendings[i], "Expected:", expected.Pendings[i])
			passed = false
		}
	}

	return passed
}

func compareSuccesses(result, expected cabby.Status, passed bool) bool {
	if result.SuccessCount != expected.SuccessCount {
		Error.Println("Got:", result.SuccessCount, "Expected:", expected.SuccessCount)
		passed = false
	}

	for i := 0; i < len(result.Successes); i++ {
		if result.Successes[i] != expected.Successes[i] {
			Error.Println("Got:", result.Successes[i], "Expected:", expected.Successes[i])
			passed = false
		}
	}
	return passed
}
