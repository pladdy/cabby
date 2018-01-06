package main

import (
	"database/sql"
	"errors"
	_ "github.com/mattn/go-sqlite3"
	"fmt"
	"io/ioutil"
	"path"
	"strings"
)

const backendDir = "backend"

type taxiiConnector interface {
	connect(connection string) error
	statement(action, name string, args map[string]string) string
	disconnect()
}

type taxiiWriter interface {
	create(name string, args map[string]string) error
}

type taxiiReader interface {
	read(name string, args map[string]string, c chan<- interface{})
}

type taxiiDataStorer interface {
	taxiiConnector
	taxiiReader
	taxiiWriter
}

type sqliteDB struct {
	db       *sql.DB
	language string
	path     string
	driver   string
}

func newTaxiiDataStore(c cabbyConfig) (taxiiDataStorer, error) {
	var t taxiiDataStorer
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
	s := sqliteDB{language: "sql", driver: "sqlite3", path: c.DataStore["path"]}
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

func (s *sqliteDB) statement(action, name string, args map[string]string) string {
  stmt := parseStatement(s.language, action, name)
	stmt = swapArgs(stmt, args)
	return stmt
}

/* writer methods */

func (s *sqliteDB) create(name string, args map[string]string) error {
	statement := s.statement("create", name, args)

	_, err := s.db.Exec(statement)
	if err != nil {
		logError.Printf("Error: '%v', in statement \"%v\"\n", err, statement)
	}
	return err
}

/* read methods */

// read methods take a name and args like create, but also a channel.  The channel is used to put returned
// data into.  It allows the caller to buffer a channel and apply back pressure.
func (s *sqliteDB) read(name string, args map[string]string, r chan<- interface{}) {
	statement := s.statement("read", name, args)

	rows, err := s.db.Query(statement)
	if err != nil {
		r <- fmt.Errorf("%v in statement: %v", err, statement)
		close(r)
		return
	}
	defer rows.Close()

  if name == "taxii_collection" {
		s.readCollection(rows, r)
	}
}

func (s *sqliteDB) readCollection(rows *sql.Rows, r chan<- interface{}) {
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

/* helpers */

func parseStatement(language, command, name string) string {
	fileName := strings.Join([]string{command, name}, "_")
	fileName = fileName + "." + language
	path := path.Join(backendDir, language, fileName)

	statement, err := ioutil.ReadFile(path)
	if err != nil {
		logError.Panicf("Unable to parse statement file: %v", path)
	}

	return string(statement)
}

func swapArgs(statement string, args map[string]string) string {
	for k, v := range args {
		statement = strings.Replace(statement, "$"+k, v, -1)
	}
	return statement
}
