package sqlite

import (
	"context"
	"database/sql"
	"strings"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	cabby "github.com/pladdy/cabby2"
)

// APIRootService implements a SQLite version of the APIRootService interface
type APIRootService struct {
	DB *sql.DB
}

// APIRoot will read from the data store and return the resource
func (s APIRootService) APIRoot(ctx context.Context, path string) (cabby.APIRoot, error) {
	resource, action := "APIRoot", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.apiRoot(path)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s APIRootService) apiRoot(path string) (cabby.APIRoot, error) {
	sql := `select api_root_path, title, description, versions, max_content_length
				  from taxii_api_root
				  where api_root_path = ?`

	a := cabby.APIRoot{}

	rows, err := s.DB.Query(sql, path)
	if err != nil {
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
		return a, err
	}
	defer rows.Close()

	for rows.Next() {
		var versions string
		if err := rows.Scan(&a.Path, &a.Title, &a.Description, &versions, &a.MaxContentLength); err != nil {
			return a, err
		}
		a.Versions = strings.Split(versions, ",")
	}

	err = rows.Err()
	return a, err
}

// APIRoots will read from the data store and return the resource
func (s APIRootService) APIRoots(ctx context.Context) ([]cabby.APIRoot, error) {
	resource, action := "APIRoots", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.apiRoots()
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s APIRootService) apiRoots() ([]cabby.APIRoot, error) {
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
	return as, err
}
