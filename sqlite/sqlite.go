package sqlite

import (
	"database/sql"
	"errors"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
)

const (
	maxWritesPerBatch = 500
)

// DataStore represents a SQLite database
type DataStore struct {
	DB   *sql.DB
	Path string
}

// NewDataStore returns a sqliteDB
func NewDataStore(path string) (*DataStore, error) {
	s := DataStore{Path: path}
	if s.Path == "" {
		return &s, errors.New("No database location specfied in config")
	}

	err := s.Open()
	return &s, err
}

// APIRootService returns a service for api root resources
func (s *DataStore) APIRootService() cabby.APIRootService {
	return APIRootService{DB: s.DB}
}

// Close connection to datastore
func (s *DataStore) Close() {
	s.DB.Close()
}

// CollectionService returns a service for collection resources
func (s *DataStore) CollectionService() cabby.CollectionService {
	return CollectionService{DB: s.DB}
}

// DiscoveryService returns a service for discovery resources
func (s *DataStore) DiscoveryService() cabby.DiscoveryService {
	return DiscoveryService{DB: s.DB}
}

// ManifestService returns a service for object resources
func (s *DataStore) ManifestService() cabby.ManifestService {
	return ManifestService{DB: s.DB}
}

// ObjectService returns a service for object resources
func (s *DataStore) ObjectService() cabby.ObjectService {
	return ObjectService{DB: s.DB, DataStore: s}
}

// Open connection to datastore
func (s *DataStore) Open() (err error) {
	// set foreign key pragma to true in connection: https://github.com/mattn/go-sqlite3#connection-string
	s.DB, err = sql.Open("sqlite3", s.Path+"?_fk=true")
	if err != nil {
		log.Error(err)
	}
	return
}

// StatusService returns service for status resources
func (s *DataStore) StatusService() cabby.StatusService {
	return StatusService{DB: s.DB, DataStore: s}
}

// UserService returns a service for user resources
func (s *DataStore) UserService() cabby.UserService {
	return UserService{DB: s.DB}
}

/* writer methods */

func (s *DataStore) batchWrite(query string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tx, stmt, err := s.writeOperation(query)
	if err != nil {
		errs <- err
		return
	}
	defer stmt.Close()

	i := 0
	for item := range toWrite {
		args := item.([]interface{})

		err := s.execute(stmt, args...)
		if err != nil {
			errs <- err
			continue
		}

		i++
		if i >= maxWritesPerBatch {
			tx.Commit() // on commit a statement is closed, create a new transaction for next batch
			tx, stmt, err = s.writeOperation(query)
			if err != nil {
				errs <- err
				return
			}
			i = 0
		}
	}
	tx.Commit()
}

func (s *DataStore) write(query string, args ...interface{}) error {
	tx, stmt, err := s.writeOperation(query)
	if err != nil {
		return err
	}
	defer tx.Commit()
	defer stmt.Close()

	return s.execute(stmt, args...)
}

func (s *DataStore) execute(stmt *sql.Stmt, args ...interface{}) error {
	_, err := stmt.Exec(args...)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "args": args}).Error("Failed to execute")
	}
	return err
}

func (s *DataStore) writeOperation(query string) (tx *sql.Tx, stmt *sql.Stmt, err error) {
	tx, err = s.DB.Begin()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to begin transaction")
		return
	}

	stmt, err = tx.Prepare(query)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "query": query}).Error("Failed to prepare query")
	}
	return
}
