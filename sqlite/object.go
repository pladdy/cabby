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
func (s ObjectService) Object(collectionID, objectID string, f cabby.Filter) ([]cabby.Object, error) {
	resource, action := "Object", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.object(collectionID, objectID, f)
	cabby.LogServiceEnd(resource, action, start)
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
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
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
func (s ObjectService) Objects(collectionID string, cr *cabby.Range, f cabby.Filter) ([]cabby.Object, error) {
	resource, action := "Objects", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.objects(collectionID, cr, f)
	cabby.LogServiceEnd(resource, action, start)
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
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
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

func updateStatus(st cabby.Status, errs chan error, ss cabby.StatusService) {
	failures := int64(0)
	for _ = range errs {
		failures++
	}

	st.FailureCount = failures
	st.SuccessCount = st.TotalCount - failures
	ss.UpdateStatus(st)
}
