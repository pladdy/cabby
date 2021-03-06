package cabby

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gofrs/uuid"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	cabbyTaxiiNamespace = "15e011d3-bcec-4f41-92d0-c6fc22ab9e45"

	// DefaultDevelopmentConfig is the path to the local dev config
	DefaultDevelopmentConfig = "config/cabby.json"
	// DefaultProductionConfig is the path to the packaged config file
	DefaultProductionConfig = "/etc/cabby/cabby.json"

	// StixContentType20 represents a stix 2.0 content type
	StixContentType20 = "application/vnd.oasis.stix+json;version=2.0"
	// StixContentType represents a stix 2 content type
	StixContentType = "application/vnd.oasis.stix+json"
	// TaxiiContentType21 represents a taxii 2.1 content type
	TaxiiContentType21 = "application/vnd.oasis.taxii+json;version=2.1"
	// TaxiiContentType represents a taxii 2 content type
	TaxiiContentType = "application/vnd.oasis.taxii+json"
	// TaxiiVersion notes the supported version of the server
	TaxiiVersion = "taxii-2.1"

	// UnsetUnixNano is the value returned from an unset time.Time{}.UnixNano() call
	UnsetUnixNano = -6795364578871345152
)

// APIRoot resource
type APIRoot struct {
	Path             string   `json:"path,omitempty"`
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

// IncludesMinVersion checks if minimum taxii version is included in list
func (a *APIRoot) IncludesMinVersion(vs []string) bool {
	versions := map[string]bool{}

	for _, v := range vs {
		versions[v] = true
	}

	b, _ := versions[TaxiiVersion]
	return b
}

// Validate an API Root
func (a *APIRoot) Validate() error {
	if a.Path == "" {
		return errors.New("Path must be defined")
	}
	if a.Title == "" {
		return errors.New("Title must be defined")
	}
	if len(a.Versions) <= 0 {
		return errors.New("At least one version must be specified")
	}
	if !a.IncludesMinVersion(a.Versions) {
		return fmt.Errorf("Minimum TAXII version %s must be included in 'Versions'", TaxiiVersion)
	}

	return nil
}

// APIRootService for interacting with APIRoots
type APIRootService interface {
	APIRoot(ctx context.Context, path string) (APIRoot, error)
	APIRoots(ctx context.Context) ([]APIRoot, error)
	CreateAPIRoot(ctx context.Context, a APIRoot) error
	DeleteAPIRoot(ctx context.Context, path string) error
	UpdateAPIRoot(ctx context.Context, a APIRoot) error
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

// Validate a collection
func (c *Collection) Validate() (err error) {
	if c.ID.IsEmpty() {
		return fmt.Errorf("Invalid id: %s", c.ID.String())
	}

	if len(c.Title) == 0 {
		return fmt.Errorf("Invalid title: %s", c.Title)
	}

	return
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
	Collection(ctx context.Context, apiRoot, collectionID string) (Collection, error)
	Collections(ctx context.Context, apiRoot string, cr *Page) (Collections, error)
	CollectionsInAPIRoot(ctx context.Context, apiRoot string) (CollectionsInAPIRoot, error)
	CreateCollection(ctx context.Context, c Collection) error
	DeleteCollection(ctx context.Context, collectionID string) error
	UpdateCollection(ctx context.Context, c Collection) error
}

// Config for a server
type Config struct {
	Host      string
	Port      int
	SSLCert   string            `json:"ssl_cert"`
	SSLKey    string            `json:"ssl_key"`
	DataStore map[string]string `json:"data_store"`
}

// Parse takes a path to a config file and converts to Configs
/* #nosec G304 */
func (c Config) Parse(file string) (initializedConfig Config) {
	if _, err := os.Stat(file); os.IsNotExist(err) {
		log.WithFields(log.Fields{"file": file, "error": err}).Panic("File does not exist")
	}

	b, err := ioutil.ReadFile(file)
	if err != nil {
		log.WithFields(log.Fields{"file": file, "error": err}).Panic("Can't parse config file")
	}

	if err = json.Unmarshal(b, &initializedConfig); err != nil {
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
	MigrationService() MigrationService
	ObjectService() ObjectService
	Open() error
	StatusService() StatusService
	UserService() UserService
	VersionsService() VersionsService
}

// Discovery resource
type Discovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

// Validate a discovery resource
func (d *Discovery) Validate() error {
	if d.Title == "" {
		return errors.New("Title must be defined")
	}

	return nil
}

// DiscoveryService interface for interacting with Discovery resources
type DiscoveryService interface {
	CreateDiscovery(ctx context.Context, d Discovery) error
	DeleteDiscovery(ctx context.Context) error
	Discovery(ctx context.Context) (Discovery, error)
	UpdateDiscovery(ctx context.Context, d Discovery) error
}

// Envelope resource for transmitting stix objects
type Envelope struct {
	More    bool              `json:"more"`
	Objects []json.RawMessage `json:"objects"`
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

// Filter for filtering results based on URL parameters
type Filter struct {
	AddedAfter stones.Timestamp
	IDs        string
	Types      string
	Versions   string
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
	ID         string           `json:"id"`
	DateAdded  stones.Timestamp `json:"date_added"`
	Version    stones.Timestamp `json:"version"`
	MediaTypes []string         `json:"media_types"`
}

// ManifestService provides manifest data
type ManifestService interface {
	Manifest(ctx context.Context, collectionID string, cr *Page, f Filter) (Manifest, error)
}

// MigrationService for performing database migrations
type MigrationService interface {
	CurrentVersion() (int, error)
	Up() error
}

// ObjectService provides Object data
type ObjectService interface {
	CreateEnvelope(ctx context.Context, e Envelope, collectionID string, s Status, ss StatusService)
	CreateObject(ctx context.Context, collectionID string, o stones.Object) error
	DeleteObject(ctx context.Context, collectionID, objecteID string) error
	Object(ctx context.Context, collectionID, objectID string, f Filter) ([]stones.Object, error)
	Objects(ctx context.Context, collectionID string, cr *Page, f Filter) ([]stones.Object, error)
}

// Page is used for paginated requests to represent the requested data range
type Page struct {
	Limit uint64
	// Used for setting X-TAXII-Date-Added-First
	MinimumAddedAfter stones.Timestamp
	// Used for setting X-TAXII-Date-Added-Last
	MaximumAddedAfter stones.Timestamp
	Total             uint64
}

// NewPage returns a Page given a string from the 'Page' HTTP header string
// the Page HTTP Header is specified by the request with the syntax 'items X-Y'
func NewPage(limit string) (p Page, err error) {
	if limit == "" {
		return p, err
	}

	limit = strings.TrimSpace(limit)

	p.Limit, err = strconv.ParseUint(limit, 10, 64)
	if err != nil {
		return p, err
	}

	if p.Valid() {
		return p, err
	}
	return p, errors.New("Invalid limit specified")
}

// AddedAfterFirst returns the first added after as a string
func (p *Page) AddedAfterFirst() string {
	return p.MinimumAddedAfter.String()
}

// AddedAfterLast returns the last added after as a string
func (p *Page) AddedAfterLast() string {
	return p.MaximumAddedAfter.String()
}

// SetAddedAfters only takes one date string and uses it to update the minimum and maximum added after fields
func (p *Page) SetAddedAfters(date string) {
	t, err := stones.TimestampFromString(date)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to parse date")
		return
	}

	if p.MinimumAddedAfter.UnixNano() == UnsetUnixNano || t.UnixNano() < p.MinimumAddedAfter.UnixNano() {
		p.MinimumAddedAfter = t
	}

	if t.UnixNano() > p.MaximumAddedAfter.UnixNano() {
		p.MaximumAddedAfter = t
	}
}

// Valid returns whether the page is valid or not
func (p *Page) Valid() bool {
	if p.Limit > 0 {
		return true
	}
	return false
}

// Status represents a TAXII status object
type Status struct {
	ID               ID               `json:"id"`
	Status           string           `json:"status"`
	RequestTimestamp stones.Timestamp `json:"request_timestamp"`
	TotalCount       int64            `json:"total_count"`
	SuccessCount     int64            `json:"success_count"`
	Successes        []string         `json:"successes"`
	FailureCount     int64            `json:"failure_count"`
	Failures         []string         `json:"failures"`
	PendingCount     int64            `json:"pending_count"`
	Pendings         []string         `json:"pendings"`
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
	CreateStatus(ctx context.Context, s Status) error
	Status(ctx context.Context, statusID string) (Status, error)
	UpdateStatus(ctx context.Context, s Status) error
}

// User represents a cabby user
// should User and UserCollectionList be combined?
type User struct {
	Email                string `json:"email"`
	CanAdmin             bool   `json:"can_admin"`
	CollectionAccessList map[ID]CollectionAccess
}

// Defined returns a bool indicating if a user is defined
func (u *User) Defined() bool {
	if u.Email == "" {
		return false
	}
	return true
}

// Validate returns whether the object is valid or not
func (u *User) Validate() (err error) {
	// Validate domain? http://data.iana.org/TLD/tlds-alpha-by-domain.txt
	re := regexp.MustCompile(`.+.@..+\...+`)
	if !re.Match([]byte(u.Email)) {
		err = fmt.Errorf("Invalid e-mail: %s", u.Email)
	}
	return
}

// UserCollectionList holds a list of collections a user can access
type UserCollectionList struct {
	Email                string                  `json:"email"`
	CollectionAccessList map[ID]CollectionAccess `json:"collection_access_list"`
}

// UserService provides Users behavior
type UserService interface {
	CreateUser(ctx context.Context, u User, password string) error
	DeleteUser(ctx context.Context, u string) error
	UpdateUser(ctx context.Context, u User) error
	CreateUserCollection(ctx context.Context, u string, ca CollectionAccess) error
	DeleteUserCollection(ctx context.Context, u, id string) error
	UpdateUserCollection(ctx context.Context, u string, ca CollectionAccess) error
	User(ctx context.Context, user, password string) (User, error)
	UserCollections(ctx context.Context, user string) (UserCollectionList, error)
}

// Versions contains a list of versions for an object
type Versions struct {
	Versions []string `json:"versions"`
}

// VersionsService provides object versions
type VersionsService interface {
	Versions(c context.Context, cid, oid string, cr *Page, f Filter) (Versions, error)
}
