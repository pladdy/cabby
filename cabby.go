package cabby

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	cabbyTaxiiNamespace = "15e011d3-bcec-4f41-92d0-c6fc22ab9e45"

	// StixContentType20 represents a stix 2.0 content type
	StixContentType20 = "application/vnd.oasis.stix+json; version=2.0"
	// StixContentType represents a stix 2 content type
	StixContentType = "application/vnd.oasis.stix+json"
	// TaxiiContentType20 represents a taxii 2.0 content type
	TaxiiContentType20 = "application/vnd.oasis.taxii+json; version=2.0"
	// TaxiiContentType represents a taxii 2 content type
	TaxiiContentType = "application/vnd.oasis.taxii+json"
)

// APIRoot resource
type APIRoot struct {
	Path             string   `json:"path,omitempty"`
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

// APIRootService for interacting with APIRoots
type APIRootService interface {
	APIRoot(path string) (APIRoot, error)
	APIRoots() ([]APIRoot, error)
}

// Collection resource
type Collection struct {
	APIRootPath string   `json:"api_root_path,omitempty"`
	ID          ID       `json:"id"`
	CanRead     bool     `json:"can_read"`
	CanWrite    bool     `json:"can_write"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"`
}

// NewCollection returns a collection resource; it takes an optional id string
func NewCollection(id ...string) (Collection, error) {
	var err error
	c := Collection{}

	// create an ID unless the parameter is a string of 'collections'...
	// TODO: document why this is here?  when can this happen and why?
	if id[0] != "collections" {
		c.ID, err = IDFromString(id[0])
	} else {
		c.ID, err = NewID()
	}

	c.MediaTypes = []string{TaxiiContentType}
	return c, err
}

// CollectionAccess defines read/write access on a collection
type CollectionAccess struct {
	ID       ID   `json:"id"`
	CanRead  bool `json:"can_read"`
	CanWrite bool `json:"can_write"`
}

// Collections resource
type Collections struct {
	Collections []Collection `json:"collections"`
}

// CollectionsInAPIRoot associated a list of collection IDs that belong to a API Root Path
type CollectionsInAPIRoot struct {
	Path          string
	CollectionIDs []ID
}

// CollectionService interface for interacting with data store
type CollectionService interface {
	Collection(user, apiRoot, collectionID string) (Collection, error)
	Collections(user, apiRoot string, cr *Range) (Collections, error)
	CollectionsInAPIRoot(apiRoot string) (CollectionsInAPIRoot, error)
}

// Config for a server
type Config struct {
	Host      string
	Port      int
	SSLCert   string            `json:"ssl_cert"`
	SSLKey    string            `json:"ssl_key"`
	DataStore map[string]string `json:"data_store"`
}

// Configs holds Configs by key (environment)
type Configs map[string]Config

// Parse takes a path to a config file and converts to Configs
func (c Configs) Parse(file string) (cs Configs) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.WithFields(log.Fields{"file": file, "error": err}).Panic("Can't parse config file")
	}

	if err = json.Unmarshal(b, &cs); err != nil {
		log.WithFields(log.Fields{"file": file, "error": err}).Panic("Can't unmarshal JSON")
	}

	return
}

// DataStore interface for backend implementations
type DataStore interface {
	APIRootService() APIRootService
	Close()
	CollectionService() CollectionService
	DiscoveryService() DiscoveryService
	ManifestService() ManifestService
	ObjectService() ObjectService
	Open() error
	StatusService() StatusService
	UserService() UserService
}

// Discovery resource
type Discovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

// DiscoveryService interface for interacting with Discovery resources
type DiscoveryService interface {
	Discovery() (Discovery, error)
}

// Error struct for TAXII 2 errors
type Error struct {
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	ErrorID         string            `json:"error_id,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	HTTPStatus      int               `json:"http_status,string,omitempty"`
	ExternalDetails string            `json:"external_details,omitempty"`
	Details         map[string]string `json:"details,omitempty"`
}

// ID for taxii resources
type ID struct {
	uuid.UUID
}

// NewID returns a new ID which is a UUID v4
func NewID() (ID, error) {
	id, err := uuid.NewV4()
	return ID{id}, err
}

// IDFromString takes a uuid string and coerces to ID
func IDFromString(s string) (ID, error) {
	id, err := uuid.FromString(s)
	return ID{id}, err
}

// IDUsingString creates a V5 UUID from the given string
func IDUsingString(s string) (ID, error) {
	ns, err := uuid.FromString(cabbyTaxiiNamespace)
	if err != nil {
		return ID{}, err
	}

	id := uuid.NewV5(ns, s)
	return ID{id}, err
}

// IsEmpty returns a boolean based on whether the UUID is not defined
//  IE: string representation 00000000-0000-0000-0000-000000000000 is undefined
func (id *ID) IsEmpty() bool {
	empty := &ID{}
	if id.String() == empty.String() {
		return true
	}
	return false
}

// Manifest resource lists a summary of objects in a collection
type Manifest struct {
	Objects []ManifestEntry `json:"objects,omitempty"`
}

// ManifestEntry is a summary of an object in a manifest
type ManifestEntry struct {
	ID         string   `json:"id"`
	DateAdded  string   `json:"date_added"`
	Versions   []string `json:"versions"`
	MediaTypes []string `json:"media_types"`
}

// ManifestService provides manifest data
type ManifestService interface {
	Manifest(collectionID string) (Manifest, error)
}

// Object for STIX 2 object data
// TODO: this should be in stones; needs validation too
type Object struct {
	ID           stones.ID `json:"id"`
	Type         string    `json:"type"`
	Created      string    `json:"created"`
	Modified     string    `json:"modified"`
	Object       []byte
	CollectionID ID
}

// ObjectService provides Object data
type ObjectService interface {
	CreateBundle(b stones.Bundle, collectionID string, s Status, ss StatusService)
	CreateObject(Object) error
	Object(collectionID, objectID string) (Object, error)
	Objects(collectionID string) ([]Object, error)
}

// Range is used for paginated requests to represent the requested data range
type Range struct {
	First int64
	Last  int64
	Total int64
}

// NewRange returns a Range given a string from the 'Range' HTTP header string
// the Range HTTP Header is specified by the request with the syntax 'items X-Y'
func NewRange(items string) (r Range, err error) {
	r = Range{First: -1, Last: -1}

	if items == "" {
		return r, err
	}

	itemDelimiter := "-"
	raw := strings.TrimSpace(items)
	tokens := strings.Split(raw, itemDelimiter)

	if len(tokens) == 2 {
		r.First, err = strconv.ParseInt(tokens[0], 10, 64)
		r.Last, err = strconv.ParseInt(tokens[1], 10, 64)
	}

	if r.Valid() {
		return r, err
	}
	return r, errors.New("Invalid range specified")
}

func (r *Range) String() string {
	s := "items " +
		strconv.FormatInt(r.First, 10) +
		"-" +
		strconv.FormatInt(r.Last, 10)

	if r.Total > 0 {
		s += "/" + strconv.FormatInt(r.Total, 10)
	}

	return s
}

// Valid returns whether the range is valid or not
func (r *Range) Valid() bool {
	if r.First < 0 || r.Last < 0 {
		return false
	}

	if r.First > r.Last {
		return false
	}

	return true
}

// Status represents a TAXII status object
type Status struct {
	ID               ID       `json:"id"`
	Status           string   `json:"status"`
	RequestTimestamp string   `json:"request_timestamp"`
	TotalCount       int64    `json:"total_count"`
	SuccessCount     int64    `json:"success_count"`
	Successes        []string `json:"successes"`
	FailureCount     int64    `json:"failure_count"`
	Failures         []string `json:"failures"`
	PendingCount     int64    `json:"pending_count"`
	Pendings         []string `json:"pendings"`
}

// NewStatus returns a status struct
func NewStatus(objects int) (Status, error) {
	if objects < 1 {
		return Status{}, errors.New("Can't post less than 1 object")
	}

	id, err := NewID()
	if err != nil {
		return Status{}, err
	}

	count := int64(objects)
	return Status{ID: id, Status: "pending", TotalCount: count, PendingCount: count}, err
}

// StatusService for status structs
type StatusService interface {
	CreateStatus(Status) error
	Status(statusID string) (Status, error)
	UpdateStatus(Status) error
}

// User represents a cabby user
// should User and UserCollectionList be combined?
type User struct {
	Email                string `json:"email"`
	CanAdmin             bool   `json:"can_admin"`
	CollectionAccessList map[ID]CollectionAccess
}

// UserCollectionList holds a list of collections a user can access
type UserCollectionList struct {
	Email                string                  `json:"email"`
	CollectionAccessList map[ID]CollectionAccess `json:"collection_access_list"`
}

// UserService provides Users behavior
type UserService interface {
	UserCollections(user string) (UserCollectionList, error)
	User(user, password string) (User, error)
	Exists(User) bool
}
