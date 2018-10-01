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

		log.WithFields(log.Fields{"id": o.ID}).Info("Sending to data store")
		toWrite <- []interface{}{o.ID, o.Type, o.Created, o.Modified, o.Object, collectionID}
	}
	close(toWrite)

	updateStatus(ctx, st, errs, ss)
}

// CreateObject will read from the data store and return the resource
func (s ObjectService) CreateObject(ctx context.Context, object cabby.Object) error {
	resource, action := "Object", "create"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.createObject(object)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s ObjectService) createObject(o cabby.Object) error {
	sql := createObjectSQL
	args := []interface{}{o.ID, o.Type, o.Created, o.Modified, o.Object, o.CollectionID}

	err := s.DataStore.write(createObjectSQL, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// Object will read from the data store and return the resource
func (s ObjectService) Object(ctx context.Context, collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
	resource, action := "Object", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.object(collectionID, objectID, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s ObjectService) object(collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
	sql := `select id, type, created, modified, object, collection_id
	        from stix_objects_data
					where
					  collection_id = ?
						and id = ?
						and $filter`

	args := []interface{}{collectionID, objectID}
	sql, args = applyFiltering(sql, f, args)

	objects := []cabby.Object{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return objects, err
	}
	defer rows.Close()

	for rows.Next() {
		var o cabby.Object
		if err := rows.Scan(&o.ID, &o.Type, &o.Created, &o.Modified, &o.Object, &o.CollectionID); err != nil {
			return objects, err
		}
		objects = append(objects, o)
	}

	err = rows.Err()
	return objects, err
}

// Objects will read from the data store and return the resource
func (s ObjectService) Objects(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
	resource, action := "Objects", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.objects(collectionID, cr, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s ObjectService) objects(collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
	sql := `with data as (
						select rowid, id, type, created, modified, object, collection_id, 1 count
						from stix_objects_data
						where
							collection_id = ?
							and $filter
					)
					select id, type, created, modified, object, collection_id, (select sum(count) from data) total
					from data
					$paginate`

	args := []interface{}{collectionID}

	sql, args = applyFiltering(sql, f, args)
	sql, args = applyPaging(sql, cr, args)

	objects := []cabby.Object{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return objects, err
	}
	defer rows.Close()

	for rows.Next() {
		var o cabby.Object
		if err := rows.Scan(&o.ID, &o.Type, &o.Created, &o.Modified, &o.Object, &o.CollectionID, &cr.Total); err != nil {
			return objects, err
		}
		objects = append(objects, o)
	}

	err = rows.Err()
	return objects, err
}

/* helpers */

func bytesToObject(b []byte) (cabby.Object, error) {
	var o cabby.Object
	err := json.Unmarshal(b, &o)
	if err != nil {
		return o, err
	}

	if o.ID == "" {
		err = errors.New("Invalid ID")
	}

	o.Object = b
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
