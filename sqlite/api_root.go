package sqlite

import (
	"database/sql"
	"strings"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// APIRootService implements a SQLite version of the APIRootService interface
type APIRootService struct {
	DB *sql.DB
}

// APIRoot will read from the data store and populate the result with a resource
func (s APIRootService) APIRoot(path string) (cabby.APIRoot, error) {
	resource, action := "api_root", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `select api_root_path, title, description, versions, max_content_length
				  from taxii_api_root
				  where api_root_path = ?`

	a := cabby.APIRoot{}

	rows, err := s.DB.Query(sql, path)
	if err != nil {
		return a, err
	}

	for rows.Next() {
		var versions string
		if err := rows.Scan(&a.Path, &a.Title, &a.Description, &versions, &a.MaxContentLength); err != nil {
			return a, err
		}
		a.Versions = strings.Split(versions, ",")
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return a, err
}

// APIRoots will read from the data store and populate the result with a resource
func (s APIRootService) APIRoots() ([]cabby.APIRoot, error) {
	resource, action := "api_roots", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `select api_root_path, title, description, versions, max_content_length
				  from taxii_api_root`

	as := []cabby.APIRoot{}

	rows, err := s.DB.Query(sql)
	if err != nil {
		return as, err
	}

	for rows.Next() {
		var a cabby.APIRoot
		var versions string

		if err := rows.Scan(&a.Path, &a.Title, &a.Description, &versions, &a.MaxContentLength); err != nil {
			return as, err
		}
		a.Versions = strings.Split(versions, ",")

		as = append(as, a)
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return as, err
}
