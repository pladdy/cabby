package main

import (
	"encoding/json"
	uuid "github.com/satori/go.uuid"
	"io/ioutil"
)

type cabbyConfig struct {
	Host       string
	Port       int
	SSLCert    string                  `json:"ssl_cert"`
	SSLKey     string                  `json:"ssl_key"`
	DataStore  map[string]string       `json:"data_store"`
	Discovery  taxiiDiscovery          `json:"discovery"`
	APIRootMap map[string]taxiiAPIRoot `json:"api_root_map"`
}

func (c *cabbyConfig) discoveryDefined() bool {
	if c.Discovery.Title == "" {
		return false
	}
	return true
}

// check if given api root is in the API Root Map of a config
func (c cabbyConfig) inAPIRootMap(ar string) (r bool) {
	for k := range c.APIRootMap {
		if ar == k {
			r = true
		}
	}
	return
}

// check if given api root is in the API Roots of a config
func (c cabbyConfig) inAPIRoots(ar string) (r bool) {
	for _, v := range c.Discovery.APIRoots {
		if ar == v {
			r = true
		}
	}
	return
}

// given a path to a config file parse it from json
func (c cabbyConfig) parse(file string) (pc cabbyConfig) {
	b, err := ioutil.ReadFile(file)
	if err != nil {
		logError.Panic(err)
	}

	if err = json.Unmarshal(b, &pc); err != nil {
		logError.Panic(err)
	}

	return
}

// validate that an api root has a same definition in apiRootMap
func (c cabbyConfig) validAPIRoot(a string) bool {
	return c.inAPIRoots(a) && c.inAPIRootMap(a)
}

type taxiiAPIRoot struct {
	Title            string   `json:"title"`
	Description      string   `json:"description,omitempty"`
	Versions         []string `json:"versions"`
	MaxContentLength int64    `json:"max_content_length"`
}

type taxiiCollection struct {
	ID          uuid.UUID `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description,omitempty"`
	CanRead     bool      `json:"can_read"`
	CanWrite    bool      `json:"can_write"`
	MediaTypes  []string  `json:"media_types,omitempty"`
}

func newTaxiiCollection(args map[string]string) (taxiiCollection, error) {
	tc := taxiiCollection{}
	uid, err := newUUID(args["id"])

	tc.ID = uid
	tc.Title = args["title"]
	tc.Description = args["description"]

	return tc, err
}

func newUUID(arg ...string) (uuid.UUID, error) {
	var err error

	if len(arg) > 0 && len(arg[0]) > 0 {
		uid, err := uuid.FromString(arg[0])
		return uid, err
	}
	return uuid.Must(uuid.NewV4()), err
}

func (c taxiiCollection) create(config cabbyConfig) error {
	t, err := newTaxiiDataStore(config)
	if err != nil {
		return err
	}

	args := map[string]string{"id": c.ID.String(), "title": c.Title, "description": c.Description}
	return t.create("taxii_collection", args)
}

func (c taxiiCollection) read(config cabbyConfig, u string, r chan<- interface{}) {
	t, err := newTaxiiDataStore(config)
	if err != nil {
		r <- err
		return
	}

	args := map[string]string{"user": u, "id": c.ID.String()}
	t.read("taxii_collection", args, r)
}

type taxiiDiscovery struct {
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	Contact     string   `json:"contact,omitempty"`
	Default     string   `json:"default,omitempty"`
	APIRoots    []string `json:"api_roots,omitempty"`
}

type taxiiError struct {
	Title           string            `json:"title"`
	Description     string            `json:"description,omitempty"`
	ErrorID         string            `json:"error_id,omitempty"`
	ErrorCode       string            `json:"error_code,omitempty"`
	HTTPStatus      int               `json:"http_status,string,omitempty"`
	ExternalDetails string            `json:"external_details,omitempty"`
	Details         map[string]string `json:"details,omitempty"`
}

type taxiiStatus struct {
	ID               string   `json:"id"`
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
