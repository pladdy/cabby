package main

import (
	"database/sql"
	"errors"
	"fmt"
	"io/ioutil"
	"path"
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
	parse(command, container string) (string, error)
}

type taxiiReader interface {
	read(query, name string, args []interface{}, results chan interface{})
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

type sqliteDB struct {
	db        *sql.DB
	dbName    string
	extension string
	path      string
	driver    string
}

func newTaxiiStorer(c cabbyConfig) (taxiiStorer, error) {
	var t taxiiStorer
	var err error

	if c.DataStore["name"] == "sqlite" {
		t, err = newSQLiteDB(c)
	} else {
		err = errors.New("Unsupported data store specified in config")
	}
	return t, err
}

/* sqlite */

func newSQLiteDB(c cabbyConfig) (*sqliteDB, error) {
	s := sqliteDB{dbName: "sqlite", extension: "sql", driver: "sqlite3", path: c.DataStore["path"]}
	if s.path == "" {
		return &s, errors.New("No database location specfied in config")
	}
	err := s.connect(s.path)
	return &s, err
}

/* connector methods */

func (s *sqliteDB) connect(connection string) error {
	logInfo.Println("Connecting to", connection)

	var err error
	s.db, err = sql.Open(s.driver, connection)
	return err
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

func (s *sqliteDB) read(query, name string, args []interface{}, r chan interface{}) {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		r <- fmt.Errorf("%v in statement: %v", err, query)
		close(r)
		return
	}
	defer rows.Close()

	if name == "taxiiCollection" {
		s.readCollection(rows, r)
	}
}

func (s *sqliteDB) readCollection(rows *sql.Rows, r chan interface{}) {
	for rows.Next() {
		var tc taxiiCollection
		var mediaTypes string

		if err := rows.Scan(&tc.ID, &tc.Title, &tc.Description, &tc.CanRead, &tc.CanWrite, &mediaTypes); err != nil {
			r <- err
			continue
		}
		tc.MediaTypes = strings.Split(mediaTypes, ",")
		r <- tc
	}

	if err := rows.Err(); err != nil {
		r <- err
	}
	close(r)
}

/* writer methods */

func (s *sqliteDB) write(query string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tx, err := s.db.Begin()
	if err != nil {
		logError.Println(err)
		errs <- err
		return
	}

	stmt, err := tx.Prepare(query)
	if err != nil {
		logError.Printf("%v in statement: %v\n", query, err)
		errs <- err
		return
	}
	defer stmt.Close()

	i := 0
	for item := range toWrite {
		args := item.([]interface{})

		_, err = stmt.Exec(args...)
		if err != nil {
			logError.Println(err)
			errs <- err
			continue
		}

		i++
		if i >= maxWrites {
			tx.Commit()
		}
	}
	tx.Commit()
}
