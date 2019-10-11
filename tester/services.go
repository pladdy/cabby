package tester

import (
	"context"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

/* DataStore */

// DataStore for tests
type DataStore struct {
	APIRootServiceFn    func() APIRootService
	CollectionServiceFn func() CollectionService
	DiscoveryServiceFn  func() DiscoveryService
	ManifestServiceFn   func() ManifestService
	MigrationServiceFn  func() MigrationService
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

// MigrationService mock
func (s DataStore) MigrationService() cabby.MigrationService {
	return s.MigrationServiceFn()
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
	APIRootFn       func(ctx context.Context, path string) (cabby.APIRoot, error)
	APIRootsFn      func(ctx context.Context) ([]cabby.APIRoot, error)
	CreateAPIRootFn func(ctx context.Context, ca cabby.APIRoot) error
	DeleteAPIRootFn func(ctx context.Context, path string) error
	UpdateAPIRootFn func(ctx context.Context, a cabby.APIRoot) error
}

// APIRoot is a mock implementation
func (s APIRootService) APIRoot(ctx context.Context, path string) (cabby.APIRoot, error) {
	return s.APIRootFn(ctx, path)
}

// APIRoots is a mock implementation
func (s APIRootService) APIRoots(ctx context.Context) ([]cabby.APIRoot, error) {
	return s.APIRootsFn(ctx)
}

// CreateAPIRoot is a mock implementation
func (s APIRootService) CreateAPIRoot(ctx context.Context, a cabby.APIRoot) error {
	return s.CreateAPIRootFn(ctx, a)
}

// DeleteAPIRoot is a mock implementation
func (s APIRootService) DeleteAPIRoot(ctx context.Context, path string) error {
	return s.DeleteAPIRootFn(ctx, path)
}

// UpdateAPIRoot is a mock implementation
func (s APIRootService) UpdateAPIRoot(ctx context.Context, a cabby.APIRoot) error {
	return s.UpdateAPIRootFn(ctx, a)
}

// CollectionService is a mock implementation
type CollectionService struct {
	CollectionFn           func(ctx context.Context, collectionID, apiRootPath string) (cabby.Collection, error)
	CollectionsFn          func(ctx context.Context, apiRootPath string, cr *cabby.Range) (cabby.Collections, error)
	CollectionsInAPIRootFn func(ctx context.Context, apiRootPath string) (cabby.CollectionsInAPIRoot, error)
	CreateCollectionFn     func(ctx context.Context, c cabby.Collection) error
	DeleteCollectionFn     func(ctx context.Context, collectionID string) error
	UpdateCollectionFn     func(ctx context.Context, c cabby.Collection) error
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

// CreateCollection is a mock implementation
func (s CollectionService) CreateCollection(ctx context.Context, c cabby.Collection) error {
	return s.CreateCollectionFn(ctx, c)
}

// DeleteCollection is a mock implementation
func (s CollectionService) DeleteCollection(ctx context.Context, id string) error {
	return s.DeleteCollectionFn(ctx, id)
}

// UpdateCollection is a mock implementation
func (s CollectionService) UpdateCollection(ctx context.Context, c cabby.Collection) error {
	return s.UpdateCollectionFn(ctx, c)
}

// DiscoveryService is a mock implementation
type DiscoveryService struct {
	CreateDiscoveryFn func(ctx context.Context, d cabby.Discovery) error
	DeleteDiscoveryFn func(ctx context.Context) error
	DiscoveryFn       func(ctx context.Context) (cabby.Discovery, error)
	UpdateDiscoveryFn func(ctx context.Context, d cabby.Discovery) error
}

// CreateDiscovery is a mock implementation
func (s DiscoveryService) CreateDiscovery(ctx context.Context, d cabby.Discovery) error {
	return s.CreateDiscoveryFn(ctx, d)
}

// DeleteDiscovery is a mock implementation
func (s DiscoveryService) DeleteDiscovery(ctx context.Context) error {
	return s.DeleteDiscoveryFn(ctx)
}

// Discovery is a mock implementation
func (s DiscoveryService) Discovery(ctx context.Context) (cabby.Discovery, error) {
	return s.DiscoveryFn(ctx)
}

// UpdateDiscovery is a mock implementation
func (s DiscoveryService) UpdateDiscovery(ctx context.Context, d cabby.Discovery) error {
	return s.UpdateDiscoveryFn(ctx, d)
}

// ManifestService is a mock implementation
type ManifestService struct {
	ManifestFn func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error)
}

// Manifest is a mock implementation
func (s ManifestService) Manifest(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
	return s.ManifestFn(ctx, collectionID, cr, f)
}

// MigrationService is a mock implementation
type MigrationService struct {
	CurrentVersionFn func() (int, error)
	UpFn             func() error
}

// CurrentVersion is a mock implementation
func (s MigrationService) CurrentVersion() (int, error) {
	return s.CurrentVersionFn()
}

// Up is a mock implementation
func (s MigrationService) Up() error {
	return s.UpFn()
}

// ObjectService is a mock implementation
type ObjectService struct {
	MaxContentLength int64
	CreateEnvelopeFn func(ctx context.Context, e cabby.Envelope, collectionID string, s cabby.Status, ss cabby.StatusService)
	CreateObjectFn   func(ctx context.Context, collectionID string, object stones.Object) error
	DeleteObjectFn   func(ctx context.Context, collectionID, objectID string) error
	ObjectFn         func(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error)
	ObjectsFn        func(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error)
}

// CreateEnvelope is a mock implementation
func (s ObjectService) CreateEnvelope(ctx context.Context, e cabby.Envelope, collectionID string, st cabby.Status, ss cabby.StatusService) {
	s.CreateEnvelopeFn(ctx, e, collectionID, st, ss)
}

// CreateObject is a mock implementation
func (s ObjectService) CreateObject(ctx context.Context, collectionID string, object stones.Object) error {
	return s.CreateObjectFn(ctx, collectionID, object)
}

// DeleteObject is a mock implementation
func (s ObjectService) DeleteObject(ctx context.Context, collectionID, objectID string) error {
	return s.DeleteObjectFn(ctx, collectionID, objectID)
}

// Object is a mock implementation
func (s ObjectService) Object(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
	return s.ObjectFn(ctx, collectionID, objectID, f)
}

// Objects is a mock implementation
func (s ObjectService) Objects(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
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
	CreateUserFn           func(ctx context.Context, u cabby.User, password string) error
	DeleteUserFn           func(ctx context.Context, u string) error
	UpdateUserFn           func(ctx context.Context, u cabby.User) error
	CreateUserCollectionFn func(ctx context.Context, u string, ca cabby.CollectionAccess) error
	DeleteUserCollectionFn func(ctx context.Context, u, id string) error
	UpdateUserCollectionFn func(ctx context.Context, u string, ca cabby.CollectionAccess) error
	UserFn                 func(ctx context.Context, user, password string) (cabby.User, error)
	UserCollectionsFn      func(ctx context.Context, user string) (cabby.UserCollectionList, error)
}

// CreateUser is a mock implementation
func (s UserService) CreateUser(ctx context.Context, user cabby.User, password string) error {
	return s.CreateUserFn(ctx, user, password)
}

// DeleteUser is a mock implementation
func (s UserService) DeleteUser(ctx context.Context, user string) error {
	return s.DeleteUserFn(ctx, user)
}

// UpdateUser is a mock implementation
func (s UserService) UpdateUser(ctx context.Context, user cabby.User) error {
	return s.UpdateUserFn(ctx, user)
}

// CreateUserCollection is a mock implementation
func (s UserService) CreateUserCollection(ctx context.Context, user string, ca cabby.CollectionAccess) error {
	return s.CreateUserCollectionFn(ctx, user, ca)
}

// DeleteUserCollection is a mock implementation
func (s UserService) DeleteUserCollection(ctx context.Context, user, id string) error {
	return s.DeleteUserCollectionFn(ctx, user, id)
}

// UpdateUserCollection is a mock implementation
func (s UserService) UpdateUserCollection(ctx context.Context, user string, ca cabby.CollectionAccess) error {
	return s.UpdateUserCollectionFn(ctx, user, ca)
}

// User is a mock implementation
func (s UserService) User(ctx context.Context, user, password string) (cabby.User, error) {
	return s.UserFn(ctx, user, password)
}

// UserCollections is a mock implementation
func (s UserService) UserCollections(ctx context.Context, user string) (cabby.UserCollectionList, error) {
	return s.UserCollectionsFn(ctx, user)
}
