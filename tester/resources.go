package tester

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
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
	ObjectID = "malware--11b940e4-4f7f-459a-80ea-9c1f17b58abc"
	// Port for testing server
	Port = 1234
	// StatusID for status tests
	StatusID = "5abf4004-4f7f-459a-2eea-9c14af7b58abc"
	// UserEmail for tests
	UserEmail = "test@cabby.com"
	// UserPassword for tests
	UserPassword = "test-password"
)

var (
	portString = strconv.Itoa(Port)

	// APIRoot mock
	APIRoot = cabby.APIRoot{
		Path:             APIRootPath,
		Title:            "test api root title",
		Description:      "test api root description",
		Versions:         []string{cabby.TaxiiVersion},
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
	// Context mock
	Context = newContext()
	// ErrorResourceNotFound mock
	ErrorResourceNotFound = cabby.Error{Title: "Resource Not Found", HTTPStatus: http.StatusNotFound}
	// Discovery mock; the handler mutates the returned path into a URL
	Discovery = cabby.Discovery{
		Title:       "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     BaseURL + "taxii/",
		APIRoots:    []string{APIRootPath}}
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
		DateAdded:  time.Now().Format(time.RFC3339Nano),
		Versions:   []string{objectCreated},
		MediaTypes: []string{cabby.StixContentType}}
	// Object mock
	Object = object()
	// Objects mock
	Objects = []stones.Object{object()}
	// Status mock
	Status = status()
	// User mock
	User = cabby.User{
		Email:                UserEmail,
		CanAdmin:             true,
		CollectionAccessList: userCollectionList().CollectionAccessList,
	}
	// UserCollectionList mock
	UserCollectionList = userCollectionList()
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

func newContext() context.Context {
	ctx := context.Background()
	ctx = cabby.WithUser(ctx, User)
	return ctx
}

func object() stones.Object {
	id, _ := stones.IdentifierFromString(ObjectID)

	o := stones.Object{
		ID:       id,
		Type:     "malware",
		Created:  objectCreated,
		Modified: objectCreated,
	}

	o.Source = []byte(`{
	      "type": "malware",
	      "id": "malware--11b940e4-4f7f-459a-80ea-9c1f17b58abc",
	      "created": "2016-04-06T20:07:09.000Z",
	      "modified": "2016-04-06T20:07:09.000Z",
	      "created_by_ref": "identity--f431f809-377b-45e0-aa1c-6a4751cae5ff",
	      "name": "Poison Ivy"
	    }`)

	return o
}

func status() cabby.Status {
	s := cabby.Status{TotalCount: 3, PendingCount: 3}
	s.ID, _ = cabby.IDFromString(StatusID)
	return s
}

func userCollectionList() cabby.UserCollectionList {
	ucl := cabby.UserCollectionList{Email: UserEmail}
	id, _ := cabby.IDFromString(CollectionID)
	ucl.CollectionAccessList = map[cabby.ID]cabby.CollectionAccess{
		id: cabby.CollectionAccess{ID: id, CanRead: true, CanWrite: true}}
	return ucl
}
