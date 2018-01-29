package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
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
	parse(command, container string) (string, error)
}

type taxiiReader interface {
	read(query, name string, args []interface{}) (interface{}, error)
}

type taxiiWriter interface {
	write(query string, toWrite chan interface{}, errors chan error)
}

type taxiiStorer interface {
	taxiiConnector
	taxiiParser
	taxiiReader
	taxiiWriter
}

func newTaxiiStorer() (t taxiiStorer, err error) {
	if config.DataStore["name"] == "sqlite" {
		t, err = newSQLiteDB()
	} else {
		err = errors.New("Unsupported data store specified in config")
	}
	return
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
	info.Println("Opening connection to", connection)
	s.db, err = sql.Open(s.driver, connection)
	return
}

func (s *sqliteDB) disconnect() {
	s.db.Close()
}

/* parser methods */

func (s *sqliteDB) parse(command, name string) (string, error) {
	path := path.Join(backendDir, s.dbName, command, name+"."+s.extension)

	statement, err := ioutil.ReadFile(path)
	if err != nil {
		err = fmt.Errorf("Unable to parse statement file: %v", path)
	}

	return string(statement), err
}

/* read methods */

func (s *sqliteDB) read(query, name string, args []interface{}) (interface{}, error) {
	var result interface{}

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return result, fmt.Errorf("%v in statement: %v", err, query)
	}

	return s.readRows(name, rows)
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

func (s *sqliteDB) readRows(name string, rows *sql.Rows) (result interface{}, err error) {
	defer rows.Close()

	switch name {
	case "taxiiCollection":
		result, err = s.readCollections(rows)
	case "taxiiCollections":
		result, err = s.readCollections(rows)
	case "taxiiUser":
		result, err = s.readUser(rows)
	default:
		err = errors.New("Unknown result name '" + name)
	}

	return
}

func (s *sqliteDB) readUser(rows *sql.Rows) (interface{}, error) {
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

/* writer methods */

func (s *sqliteDB) write(query string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tx, stmt, err := batchWriteTx(s, query, errs)
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
			tx, stmt, err = batchWriteTx(s, query, errs)
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
		fail.Printf("%v in statement: %v\n", query, err)
		errs <- err
	}

	return
}
