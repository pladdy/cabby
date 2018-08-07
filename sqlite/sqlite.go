package sqlite

import (
	"database/sql"
	"errors"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
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

// APIRootService returns a discovery service
func (s *DataStore) APIRootService() cabby.APIRootService {
	return APIRootService{DB: s.DB}
}

// Close connection to datastore
func (s *DataStore) Close() {
	s.DB.Close()
}

// DiscoveryService returns a discovery service
func (s *DataStore) DiscoveryService() cabby.DiscoveryService {
	return DiscoveryService{DB: s.DB}
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

// UserService returns a user service
func (s *DataStore) UserService() cabby.UserService {
	return UserService{DB: s.DB}
}
