package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"net/url"
	"path"
	"strconv"
	"strings"

	_ "github.com/mattn/go-sqlite3"
)

const (
	backendDir = "backend"
	maxWrites  = 500
)

type taxiiConnector interface {
	connect(connection string) error
	disconnect()
}

type taxiiParser interface {
	parse(command, resource string) (taxiiQuery, error)
}

type taxiiQuery struct {
	resource string
	query    string
}

type taxiiReader interface {
	read(resource string, args []interface{}) (interface{}, error)
}

type taxiiWriter interface {
	create(resource string, toWrite chan interface{}, errors chan error)
}

type taxiiStorer interface {
	taxiiConnector
	taxiiParser
	taxiiReader
	taxiiWriter
}

/* helpers */

func newTaxiiStorer() (t taxiiStorer, err error) {
	if config.DataStore["name"] == "sqlite" {
		t, err = newSQLiteDB()
	} else {
		err = errors.New("Unsupported data store specified in config")
	}
	return
}

func insertPort(s string) string {
	u, err := url.Parse(s)
	if err != nil {
		fail.Panic(err)
	}
	return u.Scheme + "://" + u.Host + ":" + strconv.Itoa(config.Port) + u.Path
}

/* sqlite */

type sqliteDB struct {
	db        *sql.DB
	dbName    string
	extension string
	path      string
	driver    string
}

func newSQLiteDB() (*sqliteDB, error) {
	s := sqliteDB{dbName: "sqlite", extension: "sql", driver: "sqlite3", path: config.DataStore["path"]}
	if s.path == "" {
		return &s, errors.New("No database location specfied in config")
	}
	err := s.connect(s.path)
	return &s, err
}

/* connector methods */

func (s *sqliteDB) connect(connection string) (err error) {
	s.db, err = sql.Open(s.driver, connection)
	if err != nil {
		fail.Println(err)
	}
	return
}

func (s *sqliteDB) disconnect() {
	s.db.Close()
}

/* parser methods */

func (s *sqliteDB) parse(command, resource string) (taxiiQuery, error) {
	path := path.Join(backendDir, s.dbName, command, resource+"."+s.extension)

	query, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Unable to parse file: %v", path)
	}

	return taxiiQuery{resource: resource, query: string(query)}, err
}

/* read methods */

func (s *sqliteDB) read(resource string, args []interface{}) (interface{}, error) {
	var result interface{}

	tq, err := s.parse("read", resource)
	if err != nil {
		return result, err
	}

	rows, err := s.db.Query(tq.query, args...)
	if err != nil {
		return result, fmt.Errorf("%v in statement: %v", err, tq.query)
	}

	return s.readRows(tq.resource, rows)
}

/* read helpers */

func (s *sqliteDB) readAPIRoot(rows *sql.Rows) (interface{}, error) {
	var apiRoot taxiiAPIRoot
	var err error

	for rows.Next() {
		var versions string
		if err := rows.Scan(&apiRoot.Title, &apiRoot.Description, &versions, &apiRoot.MaxContentLength); err != nil {
			return apiRoot, err
		}
		apiRoot.Versions = strings.Split(versions, ",")
	}

	err = rows.Err()
	return apiRoot, err
}

func (s *sqliteDB) readAPIRoots(rows *sql.Rows) (interface{}, error) {
	var paths []string
	var err error

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return paths, err
		}
		paths = append(paths, path)
	}

	err = rows.Err()
	return paths, err
}

func (s *sqliteDB) readCollections(rows *sql.Rows) (interface{}, error) {
	tcs := taxiiCollections{}
	var err error

	for rows.Next() {
		var tc taxiiCollection
		var mediaTypes string

		if err := rows.Scan(&tc.ID, &tc.Title, &tc.Description, &tc.CanRead, &tc.CanWrite, &mediaTypes); err != nil {
			return tcs, err
		}
		tc.MediaTypes = strings.Split(mediaTypes, ",")
		tcs.Collections = append(tcs.Collections, tc)
	}

	err = rows.Err()
	return tcs, err
}

func (s *sqliteDB) readCollectionAccess(rows *sql.Rows) (interface{}, error) {
	var tcas []taxiiCollectionAccess
	var err error

	for rows.Next() {
		var tca taxiiCollectionAccess
		if err := rows.Scan(&tca.ID, &tca.CanRead, &tca.CanWrite); err != nil {
			return tca, err
		}
		tcas = append(tcas, tca)
	}

	err = rows.Err()
	return tcas, err
}

func (s *sqliteDB) readDiscovery(rows *sql.Rows) (interface{}, error) {
	td := taxiiDiscovery{}
	var apiRoots []string
	var err error

	for rows.Next() {
		var apiRoot string
		if err := rows.Scan(&td.Title, &td.Description, &td.Contact, &td.Default, &apiRoot); err != nil {
			return td, err
		}
		td.Default = insertPort(td.Default)
		if apiRoot != "No API Roots defined" {
			apiRoots = append(apiRoots, td.Default+apiRoot)
		}
	}

	err = rows.Err()
	td.APIRoots = apiRoots
	return td, err
}

func (s *sqliteDB) readRows(resource string, rows *sql.Rows) (result interface{}, err error) {
	defer rows.Close()

	switch resource {
	case "taxiiAPIRoot":
		result, err = s.readAPIRoot(rows)
	case "taxiiAPIRoots":
		result, err = s.readAPIRoots(rows)
	case "taxiiCollection":
		result, err = s.readCollections(rows)
	case "taxiiCollections":
		result, err = s.readCollections(rows)
	case "taxiiCollectionAccess":
		result, err = s.readCollectionAccess(rows)
	case "taxiiDiscovery":
		result, err = s.readDiscovery(rows)
	case "taxiiUser":
		result, err = s.readUser(rows)
	default:
		err = errors.New("Unknown resource name '" + resource)
	}

	return
}

func (s *sqliteDB) readUser(rows *sql.Rows) (interface{}, error) {
	var valid bool
	var err error

	for rows.Next() {
		if err := rows.Scan(&valid); err != nil {
			return valid, err
		}
	}

	err = rows.Err()
	return valid, err
}

/* create methods */

func (s *sqliteDB) create(resource string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tq, err := s.parse("create", resource)
	if err != nil {
		fail.Println(err)
		errs <- err
		return
	}

	tx, stmt, err := batchWriteTx(s, tq.query, errs)
	if err != nil {
		return
	}
	defer stmt.Close()

	i := 0
	for item := range toWrite {
		args := item.([]interface{})

		_, err := stmt.Exec(args...)
		if err != nil {
			fail.Println(err)
			errs <- err
			continue
		}

		i++
		if i >= maxWrites {
			tx.Commit() // on commit a statement is closed, create a new transaction for next batch
			tx, stmt, err = batchWriteTx(s, tq.query, errs)
			if err != nil {
				return
			}
		}
	}
	tx.Commit()
}

func batchWriteTx(s *sqliteDB, query string, errs chan error) (tx *sql.Tx, stmt *sql.Stmt, err error) {
	tx, err = s.db.Begin()
	if err != nil {
		fail.Println(err)
		errs <- err
	}

	stmt, err = tx.Prepare(query)
	if err != nil {
		fail.Printf("%v in statement: %v\n", err, query)
		errs <- err
	}

	return
}
