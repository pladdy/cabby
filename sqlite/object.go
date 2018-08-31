package sqlite

import (
	"database/sql"
	"encoding/json"
	"errors"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
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
func (s ObjectService) CreateBundle(b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
	resource, action := "Bundle", "create"
	start := cabby.LogServiceStart(resource, action)
	s.createBundle(b, collectionID, st, ss)
	cabby.LogServiceEnd(resource, action, start)
}

func (s ObjectService) createBundle(b stones.Bundle, collectionID string, st cabby.Status, ss cabby.StatusService) {
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

	updateStatus(st, errs, ss)
}

// CreateObject will read from the data store and return the resource
func (s ObjectService) CreateObject(object cabby.Object) error {
	resource, action := "Object", "create"
	start := cabby.LogServiceStart(resource, action)
	err := s.createObject(object)
	cabby.LogServiceEnd(resource, action, start)
	return err
}

func (s ObjectService) createObject(o cabby.Object) error {
	return s.DataStore.write(createObjectSQL, o.ID, o.Type, o.Created, o.Modified, o.Object, o.CollectionID)
}

// Object will read from the data store and return the resource
func (s ObjectService) Object(collectionID, objectID string) (cabby.Object, error) {
	resource, action := "Object", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.object(collectionID, objectID)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s ObjectService) object(collectionID, objectID string) (cabby.Object, error) {
	sql := `select id raw_id, type, created, modified, object, collection_id
	        from stix_objects_data where collection_id = ? and id = ?`
	// add $version into query

	o := cabby.Object{}
	var err error

	rows, err := s.DB.Query(sql, collectionID, objectID)
	if err != nil {
		return o, err
	}

	for rows.Next() {
		if err := rows.Scan(&o.ID, &o.Type, &o.Created, &o.Modified, &o.Object, &o.CollectionID); err != nil {
			return o, err
		}
	}

	err = rows.Err()
	return o, err
}

// Objects will read from the data store and return the resource
func (s ObjectService) Objects(collectionID string, cr *cabby.Range) ([]cabby.Object, error) {
	resource, action := "Objects", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.objects(collectionID, cr)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s ObjectService) objects(collectionID string, cr *cabby.Range) ([]cabby.Object, error) {
	sql := `with data as (
						select rowid, id raw_id, type, created, modified, object, collection_id, 1 count
						from stix_objects_data
						where
							collection_id = ?
							/* $addedAfter
							$id
							$types
							$version */
					)
					select raw_id, type, created, modified, object, collection_id, (select sum(count) from data) total
					from data`

	var args []interface{}

	if cr.Valid() {
		sql = WithPagination(sql)
		args = []interface{}{collectionID, (cr.Last - cr.First) + 1, cr.First}
	} else {
		args = []interface{}{collectionID}
	}

	objects := []cabby.Object{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		return objects, err
	}

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

func updateStatus(st cabby.Status, errs chan error, ss cabby.StatusService) {
	failures := int64(0)
	for _ = range errs {
		failures++
	}

	st.FailureCount = failures
	ss.UpdateStatus(st)
}
