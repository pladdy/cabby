package sqlite

import (
	"context"
	"database/sql"
	"strings"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"github.com/pladdy/cabby"
)

// APIRootService implements a SQLite version of the APIRootService interface
type APIRootService struct {
	DB        *sql.DB
	DataStore *DataStore
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
				  from api_root
				  where api_root_path = ?`
	args := []interface{}{path}

	a := cabby.APIRoot{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
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
				  from api_root`
	args := []interface{}{}

	as := []cabby.APIRoot{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
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

// CreateAPIRoot creates a user in the data store
func (s APIRootService) CreateAPIRoot(ctx context.Context, a cabby.APIRoot) error {
	resource, action := "APIRoot", "create"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := a.Validate()
	if err == nil {
		err = s.createAPIRoot(a)
	} else {
		log.WithFields(log.Fields{"api_root": a, "error": err}).Error("Invalid API Root")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s APIRootService) createAPIRoot(a cabby.APIRoot) error {
	sql := `insert into api_root (api_root_path, title, description, versions, max_content_length)
					values (?, ?, ?, ?, ?)`
	args := []interface{}{a.Path, a.Title, a.Description, strings.Join(a.Versions, ","), a.MaxContentLength}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// DeleteAPIRoot creates a user in the data store
func (s APIRootService) DeleteAPIRoot(ctx context.Context, id string) error {
	resource, action := "APIRoot", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteAPIRoot(id)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s APIRootService) deleteAPIRoot(path string) error {
	sql := `delete from api_root where api_root_path = ?`
	args := []interface{}{path}

	_, err := s.DB.Exec(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// UpdateAPIRoot creates a user in the data store
func (s APIRootService) UpdateAPIRoot(ctx context.Context, a cabby.APIRoot) error {
	resource, action := "APIRoot", "update"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := a.Validate()
	if err == nil {
		err = s.updateAPIRoot(a)
	} else {
		log.WithFields(log.Fields{"api_root": a, "error": err}).Error("Invalid API Root")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s APIRootService) updateAPIRoot(a cabby.APIRoot) error {
	sql := `update api_root
				  set title = ?, description = ?, versions = ?, max_content_length = ?
				  where api_root_path = ?`
	args := []interface{}{a.Title, a.Description, strings.Join(a.Versions, ","), a.MaxContentLength, a.Path}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}
