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

// CollectionService implements a SQLite version of the CollectionService interface
type CollectionService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// Collection will read from the data store and return the resource
func (s CollectionService) Collection(ctx context.Context, apiRootPath, collectionID string) (cabby.Collection, error) {
	resource, action := "Collection", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.collection(cabby.TakeUser(ctx).Email, apiRootPath, collectionID)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s CollectionService) collection(user, apiRootPath, collectionID string) (cabby.Collection, error) {
	sql := `select c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
					from
						taxii_collection c
						inner join taxii_user_collection uc
							on c.id = uc.collection_id
					where uc.email = ? and c.api_root_path = ? and c.id = ? and uc.can_read = 1`

	c := cabby.Collection{}
	var err error

	rows, err := s.DB.Query(sql, user, apiRootPath, collectionID)
	if err != nil {
		log.WithFields(
			log.Fields{"api_root_path": apiRootPath, "collection_id": collectionID, "error": err, "sql": sql, "user": user},
		).Error("error in sql")
		return c, err
	}
	defer rows.Close()

	for rows.Next() {
		var mediaTypes string

		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CanRead, &c.CanWrite, &mediaTypes); err != nil {
			return c, err
		}
		c.MediaTypes = strings.Split(mediaTypes, ",")
	}

	err = rows.Err()
	return c, err
}

// Collections will read from the data store and return the resource
func (s CollectionService) Collections(ctx context.Context, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
	resource, action := "Collections", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.collections(cabby.TakeUser(ctx).Email, apiRootPath, cr)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s CollectionService) collections(user, apiRootPath string, cr *cabby.Range) (cabby.Collections, error) {
	sql := `with data as (
					  select id, title, description, can_read, can_write, media_types, 1 count
					  from (
						  select c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
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
					  id, title, description, can_read, can_write, media_types, (select sum(count) from data) total
				  from data
					$paginate`

	args := []interface{}{user, apiRootPath}
	sql, args = applyPaging(sql, cr, args)
	cs := cabby.Collections{}
	var err error

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		log.WithFields(
			log.Fields{"api_root_path": apiRootPath, "error": err, "sql": sql, "user": user},
		).Error("error in sql")
		return cs, err
	}
	defer rows.Close()

	for rows.Next() {
		var c cabby.Collection
		var mediaTypes string

		if err := rows.Scan(&c.ID, &c.Title, &c.Description, &c.CanRead, &c.CanWrite, &mediaTypes, &cr.Total); err != nil {
			return cs, err
		}
		c.MediaTypes = strings.Split(mediaTypes, ",")
		cs.Collections = append(cs.Collections, c)
	}

	err = rows.Err()
	return cs, err
}

// CollectionsInAPIRoot return collections in a given api root
func (s CollectionService) CollectionsInAPIRoot(ctx context.Context, apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
	resource, action := "APIRootCollections", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.collectionsInAPIRoot(apiRootPath)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s CollectionService) collectionsInAPIRoot(apiRootPath string) (cabby.CollectionsInAPIRoot, error) {
	sql := `select c.api_root_path, c.id from taxii_collection c where c.api_root_path = ?`

	ac := cabby.CollectionsInAPIRoot{}
	var err error

	rows, err := s.DB.Query(sql, apiRootPath)
	if err != nil {
		log.WithFields(log.Fields{"api_root_path": apiRootPath, "error": err, "sql": sql}).Error("error in sql")
		return ac, err
	}
	defer rows.Close()

	for rows.Next() {
		var id cabby.ID

		if err := rows.Scan(&ac.Path, &id); err != nil {
			return ac, err
		}
		ac.CollectionIDs = append(ac.CollectionIDs, id)
	}

	err = rows.Err()
	return ac, err
}

// CreateCollection creates a user in the data store
func (s CollectionService) CreateCollection(ctx context.Context, c cabby.Collection) error {
	resource, action := "Collection", "create"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := c.Validate()
	if err == nil {
		err = s.createCollection(c)
	} else {
		log.WithFields(log.Fields{"collection": c, "error": err}).Error("Invalid user and/or password")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s CollectionService) createCollection(c cabby.Collection) error {
	sql := `insert into taxii_collection (id, api_root_path, title, description, media_types)
					values (?, ?, ?, ?, ?)`

	err := s.DataStore.write(sql, c.ID.String(), c.APIRootPath, c.Title, c.Description, strings.Join(c.MediaTypes, ","))
	if err != nil {
		log.WithFields(log.Fields{"collection": c, "error": err, "sql": sql}).Error("error in sql")
	}
	return err
}

// DeleteCollection creates a user in the data store
func (s CollectionService) DeleteCollection(ctx context.Context, id string) error {
	resource, action := "Collection", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteCollection(id)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s CollectionService) deleteCollection(id string) error {
	sql := `delete from taxii_collection where id = ?`
	_, err := s.DB.Exec(sql, id)
	if err != nil {
		log.WithFields(log.Fields{"id": id, "error": err, "sql": sql}).Error("error in sql")
	}
	return err
}

// UpdateCollection creates a user in the data store
func (s CollectionService) UpdateCollection(ctx context.Context, c cabby.Collection) error {
	resource, action := "Collection", "update"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := c.Validate()
	if err == nil {
		err = s.updateCollection(c)
	} else {
		log.WithFields(log.Fields{"collection": c, "error": err}).Error("Invalid user and/or password")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s CollectionService) updateCollection(c cabby.Collection) error {
	sql := `update taxii_collection set api_root_path = ?, title = ?, description = ? where id = ?`

	err := s.DataStore.write(sql, c.APIRootPath, c.Title, c.Description, c.ID.String())
	if err != nil {
		log.WithFields(log.Fields{"collection": c, "error": err, "sql": sql}).Error("error in sql")
	}
	return err
}
