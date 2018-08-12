package tester

import (
	"strconv"

	cabby "github.com/pladdy/cabby2"
)

const (
	baseURL = "https://localhost"
	eightMB = 8388608

	// APIRootPath for tests
	APIRootPath = "cabby_test_root"
	// CollectionID for tests
	CollectionID = "82407036-edf9-4c75-9a56-e72697c53e99"
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
	// Discovery mock
	Discovery = cabby.Discovery{
		Title:       "test discovery",
		Description: "test discovery description",
		Contact:     "cabby test",
		Default:     BaseURL + "taxii/",
		APIRoots:    []string{BaseURL + APIRootPath + "/"}}
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
