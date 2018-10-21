package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	createObjectSQL = `insert into stix_objects (id, type, created, modified, object, collection_id)
				             values (?, ?, ?, ?, ?, ?)`
	batchBufferSize = 50
)

// ObjectService implements a SQLite version of the ObjectService interface
type ObjectService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// CreateBundle will read from the data store and return the resource
func (s ObjectService) CreateBundle(ctx context.Context, b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
	resource, action := "Bundle", "create"
	start := cabby.LogServiceStart(ctx, resource, action)
	s.createBundle(ctx, b, collectionID, st, ss)
	cabby.LogServiceEnd(ctx, resource, action, start)
}

func (s ObjectService) createBundle(ctx context.Context, b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
	errs := make(chan error, len(b.Objects))
	toWrite := make(chan interface{}, batchBufferSize)

	go s.DataStore.batchWrite(createObjectSQL, toWrite, errs)

	for _, object := range b.Objects {
		o, err := bytesToObject(object)
		if err != nil {
			errs <- err
			continue
		}

		log.WithFields(log.Fields{"id": o.ID.String()}).Info("Sending to data store")
		toWrite <- []interface{}{o.ID.String(), o.Type, o.Created, o.Modified, o.Source, collectionID}
	}
	close(toWrite)

	updateStatus(ctx, st, errs, ss)
}

// CreateObject will read from the data store and return the resource
func (s ObjectService) CreateObject(ctx context.Context, collectionID string, object stones.Object) error {
	resource, action := "Object", "create"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.createObject(collectionID, object)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s ObjectService) createObject(collectionID string, o stones.Object) error {
	sql := createObjectSQL
	args := []interface{}{o.ID.String(), o.Type, o.Created, o.Modified, o.Source, collectionID}

	err := s.DataStore.write(createObjectSQL, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// Object will read from the data store and return the resource
func (s ObjectService) Object(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
	resource, action := "Object", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.object(collectionID, objectID, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s ObjectService) object(collectionID, objectID string, f cabby.Filter) ([]stones.Object, error) {
	sql := `select id, type, created, modified, object
	        from stix_objects_data
					where
					  collection_id = ?
						and id = ?
						and $filter`

	args := []interface{}{collectionID, objectID}
	sql, args = applyFiltering(sql, f, args)

	objects := []stones.Object{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return objects, err
	}
	defer rows.Close()

	for rows.Next() {
		var o stones.Object
		var id string
		if err := rows.Scan(&id, &o.Type, &o.Created, &o.Modified, &o.Source); err != nil {
			return objects, err
		}

		o.ID, err = stones.IdentifierFromString(id)
		if err != nil {
			return objects, err
		}

		objects = append(objects, o)
	}

	err = rows.Err()
	return objects, err
}

// Objects will read from the data store and return the resource
func (s ObjectService) Objects(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
	resource, action := "Objects", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.objects(collectionID, cr, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s ObjectService) objects(collectionID string, cr *cabby.Range, f cabby.Filter) ([]stones.Object, error) {
	sql := `with data as (
						select rowid, id, type, created, modified, collection_id, object, created_at date_added, 1 count
						from stix_objects_data
						where
							collection_id = ?
							and $filter
					)
					select -- collection fields
					       id, type, created, modified, object,
								 -- range fields
								 date_added,
								 (select sum(count) from data) total
					from data
					$paginate`

	args := []interface{}{collectionID}

	sql, args = applyFiltering(sql, f, args)
	sql, args = applyPaging(sql, cr, args)

	objects := []stones.Object{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return objects, err
	}
	defer rows.Close()
	var dateAdded string

	for rows.Next() {
		var o stones.Object
		var id string

		if err := rows.Scan(&id, &o.Type, &o.Created, &o.Modified, &o.Source, &dateAdded, &cr.Total); err != nil {
			return objects, err
		}

		o.ID, err = stones.IdentifierFromString(id)
		if err != nil {
			return objects, err
		}

		objects = append(objects, o)
		cr.SetAddedAfters(dateAdded)
	}

	err = rows.Err()
	return objects, err
}

/* helpers */

func bytesToObject(b []byte) (stones.Object, error) {
	var o stones.Object
	err := json.Unmarshal(b, &o)
	if err != nil {
		return o, err
	}

	if valid, _ := o.ID.Valid(); !valid {
		err = errors.New("Invalid ID")
	}

	o.Source = b
	return o, err
}

func updateStatus(ctx context.Context, st cabby.Status, errs chan error, ss cabby.StatusService) {
	failures := int64(0)
	for _ = range errs {
		failures++
	}

	st.FailureCount = failures
	st.SuccessCount = st.TotalCount - failures
	ss.UpdateStatus(ctx, st)
}
