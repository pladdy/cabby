package cabby

import (
	"encoding/json"
	"io/ioutil"

	log "github.com/sirupsen/logrus"
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
	DiscoveryService() DiscoveryService
	Open() error
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

// Result struct for data returned from backend
type Result struct {
	Data      interface{}
	ItemStart int64
	ItemEnd   int64
	Items     int64
}

// User represents a cabby user
type User struct {
	Email    string `json:"email"`
	CanAdmin bool   `json:"can_admin"`
	// CollectionAccessList map[taxiiID]taxiiCollectionAccess `json:"collection_access_list"`
}

// UserService provides Users behavior
type UserService interface {
	User(user, password string) (User, error)
	Exists(User) bool
}
