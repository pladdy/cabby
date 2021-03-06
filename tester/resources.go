package tester

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	baseURL = "https://localhost"
	eightMB = 8388608

	// APIRootPath for tests
	APIRootPath = "cabby_test_root"
	// CollectionID for tests
	CollectionID = "82407036-edf9-4c75-9a56-e72697c53e99"
	// ObjectID for tests
	ObjectID = "malware--11b940e4-4f7f-459a-80ea-9c1f17b58abc"
	// Port for testing server
	Port = 1234
	// StatusID for status tests
	StatusID = "5abf4004-4f7f-459a-2eea-9c14af7b58ab"
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
	ManifestEntry = manifestEntry()
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
		CollectionAccessList: userCollectionList().CollectionAccessList}
	// UserCollectionList mock
	UserCollectionList = userCollectionList()
	// Versions mock
	Versions = cabby.Versions{
		Versions: []string{"2016-04-06T20:07:09.000Z"}}
)

func collection() cabby.Collection {
	c := cabby.Collection{
		APIRootPath: APIRootPath,
		Title:       "test collection",
		Description: "collection for testing",
		CanRead:     true,
		CanWrite:    true}

	var err error
	c.ID, err = cabby.IDFromString(CollectionID)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "input": CollectionID}).Error("Failed to create an id from provided string")
	}
	return c
}

func manifestEntry() cabby.ManifestEntry {
	now := stones.NewTimestamp()

	return cabby.ManifestEntry{
		ID:         ObjectID,
		DateAdded:  now,
		Version:    objectCreated(),
		MediaTypes: []string{cabby.StixContentType}}
}

func newContext() context.Context {
	ctx := context.Background()
	ctx = cabby.WithUser(ctx, User)
	return ctx
}

func object() (o stones.Object) {
	rawJSON := []byte(`{
	      "type": "malware",
	      "id": "malware--11b940e4-4f7f-459a-80ea-9c1f17b58abc",
	      "created": "2016-04-06T20:07:09.000Z",
	      "modified": "2016-04-06T20:07:09.000Z",
	      "created_by_ref": "identity--f431f809-377b-45e0-aa1c-6a4751cae5ff",
	      "name": "Poison Ivy"
	    }`)

	err := json.Unmarshal(rawJSON, &o)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create an object from provided string")
	}

	return
}

func objectCreated() stones.Timestamp {
	ts, err := stones.TimestampFromString("2016-04-06T20:07:09.000Z")
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to create a timestamp from provided string")
	}
	return ts
}
func status() cabby.Status {
	s := cabby.Status{TotalCount: 3, PendingCount: 3}

	var err error
	s.ID, err = cabby.IDFromString(StatusID)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "input": StatusID}).Error("Failed to create an id from provided string")
	}
	return s
}

func userCollectionList() cabby.UserCollectionList {
	ucl := cabby.UserCollectionList{Email: UserEmail}
	id, err := cabby.IDFromString(CollectionID)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "input": CollectionID}).Error("Failed to create an id from provided string")
	}

	ucl.CollectionAccessList = map[cabby.ID]cabby.CollectionAccess{
		id: cabby.CollectionAccess{ID: id, CanRead: true, CanWrite: true}}
	return ucl
}
