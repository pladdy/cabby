package sqlite

import (
	"database/sql"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// ObjectService implements a SQLite version of the ObjectService interface
type ObjectService struct {
	DB *sql.DB
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
func (s ObjectService) Objects(collectionID string) (cabby.Objects, error) {
	resource, action := "Objects", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.objects(collectionID)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s ObjectService) objects(collectionID string) (cabby.Objects, error) {
	// filtering and pagination omitted below
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
					select raw_id, type, created, modified, object, collection_id
					-- , (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
					from data`
	// add $paginate here

	objects := cabby.Objects{}
	var err error

	rows, err := s.DB.Query(sql, collectionID)
	if err != nil {
		return objects, err
	}

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
