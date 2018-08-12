package sqlite

import (
	"database/sql"
	"strings"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// CollectionService implements a SQLite version of the CollectionService interface
type CollectionService struct {
	DB *sql.DB
}

// Collection will read from the data store and populate the result with a resource
func (s CollectionService) Collection(user, apiRoot, collectionID string) (cabby.Collection, error) {
	resource, action := "Collection", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `select c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
					from
						taxii_collection c
						inner join taxii_user_collection uc
							on c.id = uc.collection_id
					where uc.email = ? and c.api_root_path = ? and c.id = ? and uc.can_read = 1`

	c := cabby.Collection{}
	var err error

	rows, err := s.DB.Query(sql, user, apiRoot, collectionID)
	if err != nil {
		return c, err
	}

	for rows.Next() {
		var mediaTypes string

		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CanRead, &c.CanWrite, &mediaTypes); err != nil {
			return c, err
		}
		c.MediaTypes = strings.Split(mediaTypes, ",")
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return c, err
}

// Collections will read from the data store and populate the result with a resource
func (s CollectionService) Collections(user, apiRoot string) (cabby.Collections, error) {
	resource, action := "Collections", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `with data as (
					  select rowid, id, title, description, can_read, can_write, media_types, 1 count
					  from (
						  select c.rowid, c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
						  from
							  taxii_collection c
							  inner join taxii_user_collection uc
								  on c.id = uc.collection_id
						  where
						 	 uc.email = ?
							 and c.api_root_path = ?
					 		 and (uc.can_read = 1 or uc.can_write = 1)
					  )
				  )
				  select
					  -- pagination omitted for now
					  id, title, description, can_read, can_write, media_types --,
					  -- (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
				  from data`
	// add $paginate here

	cs := cabby.Collections{}
	var err error

	rows, err := s.DB.Query(sql, user, apiRoot)
	if err != nil {
		return cs, err
	}

	for rows.Next() {
		var c cabby.Collection
		var mediaTypes string

		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CanRead, &c.CanWrite, &mediaTypes); err != nil {
			return cs, err
		}
		c.MediaTypes = strings.Split(mediaTypes, ",")
		cs.Collections = append(cs.Collections, c)
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return cs, err
}

// CollectionsInAPIRoot return collections in a given api root
func (s CollectionService) CollectionsInAPIRoot(apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
	resource, action := "APIRootCollections", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `select c.api_root_path, c.id
					from
						taxii_collection c
					where c.api_root_path = ?`

	ac := cabby.CollectionsInAPIRoot{}
	var err error

	rows, err := s.DB.Query(sql, apiRootPath)
	if err != nil {
		return ac, err
	}

	for rows.Next() {
		var id cabby.ID

		if err := rows.Scan(&ac.Path, &id); err != nil {
			return ac, err
		}
		ac.CollectionIDs = append(ac.CollectionIDs, id)
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return ac, err
}
