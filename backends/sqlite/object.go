package sqlite

import (
	"context"
	"database/sql"
	"encoding/json"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

const (
	createObjectSQL = `insert into objects (id, type, created, modified, object, collection_id)
				             values (?, ?, ?, ?, ?, ?)`
	deleteObjectSQL = `delete from objects where collection_id = ? and id = ?`
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

	for _, raw := range b.Objects {
		var o stones.Object
		err := json.Unmarshal(raw, &o)
		if err != nil {
			log.WithFields(log.Fields{"raw object": string(raw), "error": err}).Error("Failed to convert bytes to Object")
			errs <- err
			continue
		}

		valid, validationErrs := o.Valid()
		if !valid {
			err = stones.ErrorsToString(validationErrs)
			log.WithFields(log.Fields{"raw object": string(raw), "error": err}).Error("Invalid object")
			errs <- err
			continue
		}

		log.WithFields(log.Fields{"id": o.ID.String()}).Info("Sending to data store")
		toWrite <- []interface{}{o.ID.String(), o.Type, o.Created.String(), o.Modified.String(), o.Source, collectionID}
	}
	close(toWrite)

	updateStatus(ctx, st, errs, ss)
}

// CreateObject will create an object in the datastore
func (s ObjectService) CreateObject(ctx context.Context, collectionID string, object stones.Object) error {
	resource, action := "Object", "create"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.createObject(collectionID, object)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s ObjectService) createObject(collectionID string, o stones.Object) error {
	sql := createObjectSQL
	args := []interface{}{o.ID.String(), o.Type, o.Created.String(), o.Modified.String(), o.Source, collectionID}

	err := s.DataStore.write(createObjectSQL, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// DeleteObject will delete an object from a collection
func (s ObjectService) DeleteObject(ctx context.Context, collectionID, objectID string) error {
	resource, action := "Object", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteObject(collectionID, objectID)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s ObjectService) deleteObject(collectionID, objectID string) error {
	sql := deleteObjectSQL
	args := []interface{}{collectionID, objectID}

	err := s.DataStore.write(deleteObjectSQL, args...)
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
	        from objects_data
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
		var id, created, modified string

		if err := rows.Scan(&id, &o.Type, &created, &modified, &o.Source); err != nil {
			return objects, err
		}

		o, err = unmarshalObject(o, id, created, modified)
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
						from objects_data
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
		var id, created, modified string

		if err := rows.Scan(&id, &o.Type, &created, &modified, &o.Source, &dateAdded, &cr.Total); err != nil {
			return objects, err
		}

		o, err = unmarshalObject(o, id, created, modified)
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

func unmarshalObject(o stones.Object, id, created, modified string) (new stones.Object, err error) {
	o.ID, err = stones.IdentifierFromString(id)
	if err != nil {
		return o, err
	}

	ts, err := stones.TimestampFromString(created)
	if err != nil {
		return o, err
	}
	o.Created = ts

	ts, err = stones.TimestampFromString(modified)
	if err != nil {
		return o, err
	}
	o.Modified = ts

	new = o
	return
}

func updateStatus(ctx context.Context, st cabby.Status, errs chan error, ss cabby.StatusService) {
	for err := range errs {
		if err != nil {
			log.WithFields(log.Fields{"error": err}).Warn("Found an error")
			st.FailureCount++
		}
	}

	// don't allow more failures than objects that can be written; TODO: provide better status updates
	if st.FailureCount > st.TotalCount {
		st.FailureCount = st.TotalCount
	}

	st.SuccessCount = st.TotalCount - st.FailureCount

	err := ss.UpdateStatus(ctx, st)
	if err != nil {
		log.WithFields(log.Fields{"error": err, "status": st}).Error("An error occured when updating the status")
	}
}
