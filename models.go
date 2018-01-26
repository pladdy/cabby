package main

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"strings"

	uuid "github.com/satori/go.uuid"
)

const minBuffer = 10

/* helpers */

func processReadError(err error, results chan interface{}) {
	logError.Println(err)
	results <- err
	close(results)
}

/* models and methods */

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
	ID          taxiiID  `json:"id"`
	CanRead     bool     `json:"can_read"`
	CanWrite    bool     `json:"can_write"`
	Title       string   `json:"title"`
	Description string   `json:"description,omitempty"`
	MediaTypes  []string `json:"media_types,omitempty"`
}

func newTaxiiCollection(id ...string) (taxiiCollection, error) {
	var err error
	tc := taxiiCollection{}
	tc.ID, err = newTaxiiID(id[0])
	tc.MediaTypes = []string{stixContentType}
	return tc, err
}

// creating a collection is a two step process: create the collection then create the association of the collection
// to the api root
func (c taxiiCollection) create(user, apiRoot string) error {
	var err error

	parts := []struct {
		name string
		args []interface{}
	}{
		{"taxiiCollection", []interface{}{c.ID.String(), c.Title, c.Description, strings.Join(c.MediaTypes, ",")}},
		{"taxiiCollectionAPIRoot", []interface{}{c.ID.String(), apiRoot}},
		{"taxiiUserCollection", []interface{}{user, c.ID.String(), true, true}},
	}

	for _, p := range parts {
		err := createTaxiiCollectionPart(p.name, p.args)
		if err != nil {
			return err
		}
	}

	return err
}

func createTaxiiCollectionPart(part string, args []interface{}) error {
	t, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	query, err := t.parse("create", part)
	if err != nil {
		return err
	}

	toWrite := make(chan interface{}, minBuffer)
	errs := make(chan error, minBuffer)

	go t.write(query, toWrite, errs)
	toWrite <- args
	close(toWrite)

	for e := range errs {
		err = e
	}

	return err
}

func (c taxiiCollection) read(u string, results chan interface{}) {
	t, err := newTaxiiStorer()
	if err != nil {
		processReadError(err, results)
		return
	}

	query, err := t.parse("read", "taxiiCollection")
	if err != nil {
		processReadError(err, results)
		return
	}

	args := []interface{}{u, c.ID.String()}
	t.read(query, "taxiiCollection", args, results)
}

type taxiiCollectionAccess struct {
	ID       taxiiID `json:"id"`
	CanRead  bool    `json:"can_read"`
	CanWrite bool    `json:"can_write"`
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

type taxiiID struct {
	uuid.UUID
}

// func (t taxiiID) String() string {
// 	return t.String()
// }

func newTaxiiID(arg ...string) (taxiiID, error) {
	if len(arg) > 0 && len(arg[0]) > 0 {
		id, err := uuid.FromString(arg[0])
		return taxiiID{id}, err
	}

	id, err := uuid.NewV4()
	return taxiiID{id}, err
}

type taxiiStatus struct {
	ID               taxiiID  `json:"id"`
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

type taxiiUser struct {
	Email            string
	CollectionAccess map[taxiiID]taxiiCollectionAccess
}

func newTaxiiUser(u, p string) (taxiiUser, error) {
	tu := taxiiUser{Email: u, CollectionAccess: make(map[taxiiID]taxiiCollectionAccess)}
	results := make(chan interface{}, minBuffer)
	var err error

	go tu.read(fmt.Sprintf("%x", sha256.Sum256([]byte(p))), results)

	for r := range results {
		switch r.(type) {
		case error:
			return tu, r.(error)
		}
		a := r.(taxiiCollectionAccess)
		tu.CollectionAccess[a.ID] = a
	}

	if len(tu.CollectionAccess) <= 0 {
		err = errors.New("No access to any collections")
	}
	return tu, err
}

func (tu taxiiUser) read(pass string, results chan interface{}) {
	t, err := newTaxiiStorer()
	if err != nil {
		processReadError(err, results)
		return
	}

	query, err := t.parse("read", "taxiiUser")
	if err != nil {
		processReadError(err, results)
		return
	}

	args := []interface{}{tu.Email, pass}
	t.read(query, "taxiiUser", args, results)
}
