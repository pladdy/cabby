package tester

import (
	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/stones"
)

/* DataStore */

// DataStore for tests
type DataStore struct {
	APIRootServiceFn    func() APIRootService
	CollectionServiceFn func() CollectionService
	DiscoveryServiceFn  func() DiscoveryService
	ManifestServiceFn   func() ManifestService
	ObjectServiceFn     func() ObjectService
	StatusServiceFn     func() StatusService
	UserServiceFn       func() UserService
}

// NewDataStore structure
func NewDataStore() *DataStore {
	return &DataStore{}
}

// APIRootService mock
func (s DataStore) APIRootService() cabby.APIRootService {
	return s.APIRootServiceFn()
}

// Close mock
func (s DataStore) Close() {
	return
}

// CollectionService mock
func (s DataStore) CollectionService() cabby.CollectionService {
	return s.CollectionServiceFn()
}

// DiscoveryService mock
func (s DataStore) DiscoveryService() cabby.DiscoveryService {
	return s.DiscoveryServiceFn()
}

// ManifestService mock
func (s DataStore) ManifestService() cabby.ManifestService {
	return s.ManifestServiceFn()
}

// ObjectService mock
func (s DataStore) ObjectService() cabby.ObjectService {
	return s.ObjectServiceFn()
}

// Open mock
func (s DataStore) Open() error {
	return nil
}

// StatusService mock
func (s DataStore) StatusService() cabby.StatusService {
	return s.StatusServiceFn()
}

// UserService mock
func (s DataStore) UserService() cabby.UserService {
	return s.UserServiceFn()
}

/* services */

// APIRootService is a mock implementation
type APIRootService struct {
	APIRootFn  func(path string) (cabby.APIRoot, error)
	APIRootsFn func() ([]cabby.APIRoot, error)
}

// APIRoot is a mock implementation
func (s APIRootService) APIRoot(path string) (cabby.APIRoot, error) {
	return s.APIRootFn(path)
}

// APIRoots is a mock implementation
func (s APIRootService) APIRoots() ([]cabby.APIRoot, error) {
	return s.APIRootsFn()
}

// CollectionService is a mock implementation
type CollectionService struct {
	CollectionFn           func(user, collectionID, apiRootPath string) (cabby.Collection, error)
	CollectionsFn          func(user, apiRootPath string, cr *cabby.Range) (cabby.Collections, error)
	CollectionsInAPIRootFn func(apiRootPath string) (cabby.CollectionsInAPIRoot, error)
}

// Collection is a mock implementation
func (s CollectionService) Collection(user, collectionID, apiRootPath string) (cabby.Collection, error) {
	return s.CollectionFn(user, collectionID, apiRootPath)
}

// Collections is a mock implementation
func (s CollectionService) Collections(user, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
	return s.CollectionsFn(user, apiRootPath, cr)
}

// CollectionsInAPIRoot is a mock implementation
func (s CollectionService) CollectionsInAPIRoot(apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
	return s.CollectionsInAPIRootFn(apiRootPath)
}

// DiscoveryService is a mock implementation
type DiscoveryService struct {
	DiscoveryFn func() (cabby.Discovery, error)
}

// Discovery is a mock implementation
func (s DiscoveryService) Discovery() (cabby.Discovery, error) {
	return s.DiscoveryFn()
}

// ManifestService is a mock implementation
type ManifestService struct {
	ManifestFn func(collectionID string) (cabby.Manifest, error)
}

// Manifest is a mock implementation
func (s ManifestService) Manifest(collectionID string) (cabby.Manifest, error) {
	return s.ManifestFn(collectionID)
}

// ObjectService is a mock implementation
type ObjectService struct {
	MaxContentLength int64
	CreateBundleFn   func(b stones.Bundle, collectionID string, s cabby.Status, ss cabby.StatusService)
	CreateObjectFn   func(object cabby.Object) error
	ObjectFn         func(collectionID, objectID string) (cabby.Object, error)
	ObjectsFn        func(collectionID string) ([]cabby.Object, error)
}

// CreateBundle is a mock implementation
func (s ObjectService) CreateBundle(b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
	s.CreateBundleFn(b, collectionID, st, ss)
}

// CreateObject is a mock implementation
func (s ObjectService) CreateObject(object cabby.Object) error {
	return s.CreateObjectFn(object)
}

// Object is a mock implementation
func (s ObjectService) Object(collectionID, objectID string) (cabby.Object, error) {
	return s.ObjectFn(collectionID, objectID)
}

// Objects is a mock implementation
func (s ObjectService) Objects(collectionID string) ([]cabby.Object, error) {
	return s.ObjectsFn(collectionID)
}

// StatusService is a mock implementation
type StatusService struct {
	CreateStatusFn func(status cabby.Status) error
	StatusFn       func(statusID string) (cabby.Status, error)
	UpdateStatusFn func(status cabby.Status) error
}

// CreateStatus is a mock implementation
func (s StatusService) CreateStatus(status cabby.Status) error {
	return s.CreateStatusFn(status)
}

// Status is a mock implementation
func (s StatusService) Status(statusID string) (cabby.Status, error) {
	return s.StatusFn(statusID)
}

// UpdateStatus is a mock implementation
func (s StatusService) UpdateStatus(status cabby.Status) error {
	return s.UpdateStatusFn(status)
}

// UserService is a mock implementation
type UserService struct {
	UserFn            func(user, password string) (cabby.User, error)
	UserCollectionsFn func(user string) (cabby.UserCollectionList, error)
	ExistsFn          func(cabby.User) bool
}

// Exists is a mock implementation
func (s UserService) Exists(u cabby.User) bool {
	return s.ExistsFn(u)
}

// User is a mock implementation
func (s UserService) User(user, password string) (cabby.User, error) {
	return s.UserFn(user, password)
}

// UserCollections is a mock implementation
func (s UserService) UserCollections(user string) (cabby.UserCollectionList, error) {
	return s.UserCollectionsFn(user)
}
