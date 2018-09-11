package tester

import (
	"context"

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
	APIRootFn  func(ctx context.Context, path string) (cabby.APIRoot, error)
	APIRootsFn func(ctx context.Context) ([]cabby.APIRoot, error)
}

// APIRoot is a mock implementation
func (s APIRootService) APIRoot(ctx context.Context, path string) (cabby.APIRoot, error) {
	return s.APIRootFn(ctx, path)
}

// APIRoots is a mock implementation
func (s APIRootService) APIRoots(ctx context.Context) ([]cabby.APIRoot, error) {
	return s.APIRootsFn(ctx)
}

// CollectionService is a mock implementation
type CollectionService struct {
	CollectionFn           func(ctx context.Context, collectionID, apiRootPath string) (cabby.Collection, error)
	CollectionsFn          func(ctx context.Context, apiRootPath string, cr *cabby.Range) (cabby.Collections, error)
	CollectionsInAPIRootFn func(ctx context.Context, apiRootPath string) (cabby.CollectionsInAPIRoot, error)
}

// Collection is a mock implementation
func (s CollectionService) Collection(ctx context.Context, collectionID, apiRootPath string) (cabby.Collection, error) {
	return s.CollectionFn(ctx, collectionID, apiRootPath)
}

// Collections is a mock implementation
func (s CollectionService) Collections(ctx context.Context, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
	return s.CollectionsFn(ctx, apiRootPath, cr)
}

// CollectionsInAPIRoot is a mock implementation
func (s CollectionService) CollectionsInAPIRoot(ctx context.Context, apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
	return s.CollectionsInAPIRootFn(ctx, apiRootPath)
}

// DiscoveryService is a mock implementation
type DiscoveryService struct {
	DiscoveryFn func(ctx context.Context) (cabby.Discovery, error)
}

// Discovery is a mock implementation
func (s DiscoveryService) Discovery(ctx context.Context) (cabby.Discovery, error) {
	return s.DiscoveryFn(ctx)
}

// ManifestService is a mock implementation
type ManifestService struct {
	ManifestFn func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error)
}

// Manifest is a mock implementation
func (s ManifestService) Manifest(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
	return s.ManifestFn(ctx, collectionID, cr, f)
}

// ObjectService is a mock implementation
type ObjectService struct {
	MaxContentLength int64
	CreateBundleFn   func(ctx context.Context, b stones.Bundle, collectionID string, s cabby.Status, ss cabby.StatusService)
	CreateObjectFn   func(ctx context.Context, object cabby.Object) error
	ObjectFn         func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error)
	ObjectsFn        func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error)
}

// CreateBundle is a mock implementation
func (s ObjectService) CreateBundle(ctx context.Context, b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
	s.CreateBundleFn(ctx, b, collectionID, st, ss)
}

// CreateObject is a mock implementation
func (s ObjectService) CreateObject(ctx context.Context, object cabby.Object) error {
	return s.CreateObjectFn(ctx, object)
}

// Object is a mock implementation
func (s ObjectService) Object(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
	return s.ObjectFn(ctx, collectionID, objectID, f)
}

// Objects is a mock implementation
func (s ObjectService) Objects(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
	return s.ObjectsFn(ctx, collectionID, cr, f)
}

// StatusService is a mock implementation
type StatusService struct {
	CreateStatusFn func(ctx context.Context, status cabby.Status) error
	StatusFn       func(ctx context.Context, statusID string) (cabby.Status, error)
	UpdateStatusFn func(ctx context.Context, status cabby.Status) error
}

// CreateStatus is a mock implementation
func (s StatusService) CreateStatus(ctx context.Context, status cabby.Status) error {
	return s.CreateStatusFn(ctx, status)
}

// Status is a mock implementation
func (s StatusService) Status(ctx context.Context, statusID string) (cabby.Status, error) {
	return s.StatusFn(ctx, statusID)
}

// UpdateStatus is a mock implementation
func (s StatusService) UpdateStatus(ctx context.Context, status cabby.Status) error {
	return s.UpdateStatusFn(ctx, status)
}

// UserService is a mock implementation
type UserService struct {
	CreateUserFn      func(ctx context.Context, u cabby.User, password string) error
	DeleteUserFn      func(ctx context.Context, u string) error
	UserFn            func(ctx context.Context, user, password string) (cabby.User, error)
	UserCollectionsFn func(ctx context.Context, user string) (cabby.UserCollectionList, error)
}

// CreateUser is a mock implementation
func (s UserService) CreateUser(ctx context.Context, user cabby.User, password string) error {
	return s.CreateUserFn(ctx, user, password)
}

// DeleteUser is a mock implementation
func (s UserService) DeleteUser(ctx context.Context, user string) error {
	return s.DeleteUserFn(ctx, user)
}

// User is a mock implementation
func (s UserService) User(ctx context.Context, user, password string) (cabby.User, error) {
	return s.UserFn(ctx, user, password)
}

// UserCollections is a mock implementation
func (s UserService) UserCollections(ctx context.Context, user string) (cabby.UserCollectionList, error) {
	return s.UserCollectionsFn(ctx, user)
}
