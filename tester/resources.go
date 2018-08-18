package tester

import (
	"strconv"

	cabby "github.com/pladdy/cabby2"
)

const (
	baseURL       = "https://localhost"
	eightMB       = 8388608
	objectCreated = "2016-04-06T20:07:09.000Z"

	// APIRootPath for tests
	APIRootPath = "cabby_test_root"
	// CollectionID for tests
	CollectionID = "82407036-edf9-4c75-9a56-e72697c53e99"
	// ObjectID for tests
	ObjectID = "malware--31b940d4-6f7f-459a-80ea-9c1f17b5891b"
	// Port for testing server
	Port = 1234
	// UserEmail for tests
	UserEmail = "test@cabby.com"
	// UserPassword for tests
	UserPassword = "test"
)

var (
	portString = strconv.Itoa(Port)

	// APIRoot mock
	APIRoot = cabby.APIRoot{
		Path:             APIRootPath,
		Title:            "test api root title",
		Description:      "test api root description",
		Versions:         []string{"taxii-2.0"},
		MaxContentLength: eightMB}

	// BaseURL for tests
	BaseURL = baseURL + ":" + portString + "/"

	// Collection mock
	Collection = collection()
	// Collections mock
	Collections = cabby.Collections{
		Collections: []cabby.Collection{Collection}}
	// CollectionsInAPIRoot mock
	CollectionsInAPIRoot = cabby.CollectionsInAPIRoot{
		Path: APIRootPath, CollectionIDs: []cabby.ID{Collection.ID}}
	// Discovery mock; the handler mutates the returned path into a URL
	Discovery = cabby.Discovery{
		Title:       "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     BaseURL + "taxii/",
		APIRoots:    []string{BaseURL + APIRootPath + "/"}}
	// DiscoveryDataStore mock; the service just returns an API root path
	DiscoveryDataStore = cabby.Discovery{
		Title:       "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     BaseURL + "taxii/",
		APIRoots:    []string{APIRootPath}}
	// Manifest mock
	Manifest = cabby.Manifest{Objects: []cabby.ManifestEntry{ManifestEntry}}
	// ManifestEntry mock
	ManifestEntry = cabby.ManifestEntry{
		ID:         ObjectID,
		DateAdded:  objectCreated,
		Versions:   []string{objectCreated},
		MediaTypes: []string{cabby.StixContentType}}
	// Object mock
	Object = object()
	// Objects mock
	Objects = []cabby.Object{object()}
	// User mock
	User = cabby.User{
		Email:    UserEmail,
		CanAdmin: true}
)

func collection() cabby.Collection {
	c := cabby.Collection{
		APIRootPath: APIRootPath,
		Title:       "test collection",
		Description: "collection for testing",
		CanRead:     true,
		CanWrite:    true}

	c.ID, _ = cabby.IDFromString(CollectionID)
	return c
}

func object() cabby.Object {
	o := cabby.Object{
		ID:       ObjectID,
		Type:     "malware",
		Created:  objectCreated,
		Modified: objectCreated,
	}

	o.CollectionID, _ = cabby.IDFromString(CollectionID)
	o.Object = []byte(`{
	      "type": "malware",
	      "id": "malware--31b940d4-6f7f-459a-80ea-9c1f17b5891b",
	      "created": "2016-04-06T20:07:09.000Z",
	      "modified": "2016-04-06T20:07:09.000Z",
	      "created_by_ref": "identity--f431f809-377b-45e0-aa1c-6a4751cae5ff",
	      "name": "Poison Ivy"
	    }`)

	return o
}
