package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type sqliteDB struct {
	db        *sql.DB
	dbName    string
	extension string
	path      string
	driver    string
}

var statements = map[string]map[string]string{
	"create": map[string]string{
		"stixObject": `insert into stix_objects (id, type, created, modified, object, collection_id)
		               values (?, ?, ?, ?, ?, ?)`,
		"taxiiAPIRoot": `insert into taxii_api_root (id, api_root_path, title, description, versions, max_content_length)
		                 values (?, ?, ?, ?, ?, ?)`,
		"taxiiCollection": `insert into taxii_collection (id, api_root_path, title, description, media_types)
		                    values (?, ?, ?, ?, ?)`,
		"taxiiCollectionAPIRoot": `insert into taxii_collection_api_root (collection_id, api_root_id)
		                           values (?, ?)`,
		"taxiiDiscovery": `insert into taxii_discovery (title, description, contact, default_url)
		                   values (?, ?, ?, ?)`,
		"taxiiUser": `insert into taxii_user (email) values (?)`,
		"taxiiUserCollection": `insert into taxii_user_collection (email, collection_id, can_read, can_write)
		                        values (?, ?, ?, ?)`,
		"taxiiUserPass": `insert into taxii_user_pass (email, pass) values (?, ?)`,
	},
	"read": map[string]string{
		"routableCollections": `select id from taxii_collection where api_root_path = ?`,
		"stixObject":          `select object from stix_objects where collection_id = ? and id = ?`,
		"stixObjects":         `select object from stix_objects where collection_id = ?`,
		"taxiiAPIRoot": `select title, description, versions, max_content_length
		                 from taxii_api_root
		                 where api_root_path = ?`,
		"taxiiAPIRoots": `select api_root_path from taxii_api_root`,
		"taxiiCollection": `select c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
												from
												  taxii_collection c
												  inner join taxii_user_collection uc
												    on c.id = uc.collection_id
												where uc.email = ? and c.id = ? and uc.can_read = 1`,
		"taxiiCollectionAccess": `select tuc.collection_id, tuc.can_read, tuc.can_write
															from
															  taxii_user tu
															  inner join taxii_user_collection tuc
															    on tu.email = tuc.email
															where tu.email = ?`,
		"taxiiCollections": `select c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
												 from
												   taxii_collection c
												   inner join taxii_user_collection uc
												     on c.id = uc.collection_id
												 where
												   uc.email = ?
												   and uc.can_read = 1 or uc.can_write = 1`,
		"taxiiDiscovery": `select td.title, td.description, td.contact, td.default_url,
											   case
											     when tar.api_root_path is null then 'No API Roots defined' else tar.api_root_path
											   end api_root_path
											 from
											   taxii_discovery td
											   left join taxii_api_root tar
											     on td.id = tar.discovery_id`,
		"taxiiManifest": `select id, min(created) date_added, group_concat(modified) versions -- media_types omitted for now
											from stix_objects
											where collection_id = ?
											group by id`,
		"taxiiUser": `select 1
									from
									  taxii_user tu
									  inner join taxii_user_pass tup
									    on tu.email = tup.email
									where tu.email = ? and tup.pass = ?`,
	},
}

func newSQLiteDB(path string) (*sqliteDB, error) {
	var s sqliteDB

	s = sqliteDB{dbName: "sqlite", extension: "sql", driver: "sqlite3", path: path}
	if s.path == "" {
		return &s, errors.New("No database location specfied in config")
	}
	err := s.connect(s.path)
	return &s, err
}

/* connector methods */

func (s *sqliteDB) connect(connection string) (err error) {
	s.db, err = sql.Open(s.driver, connection)
	if err != nil {
		log.Error(err)
	}
	return
}

func (s *sqliteDB) disconnect() {
	s.db.Close()
}

/* parser methods */

func (s *sqliteDB) parse(command, resource string) (taxiiQuery, error) {
	var err error
	statement := statements[command][resource]

	if statement == "" {
		return taxiiQuery{}, fmt.Errorf("invalid command %v and resource %v", command, resource)
	}
	return taxiiQuery{resource: resource, query: statement}, err
}

/* read methods */

func (s *sqliteDB) read(resource string, args []interface{}) (interface{}, error) {
	var result interface{}

	tq, err := s.parse("read", resource)
	if err != nil {
		return result, err
	}

	rows, err := s.db.Query(tq.query, args...)
	if err != nil {
		return result, fmt.Errorf("%v in statement: %v", err, tq.query)
	}

	return s.readRows(tq.resource, rows)
}

/* read helpers */

func (s *sqliteDB) readAPIRoot(rows *sql.Rows) (interface{}, error) {
	var apiRoot taxiiAPIRoot
	var err error

	for rows.Next() {
		var versions string
		if err := rows.Scan(&apiRoot.Title, &apiRoot.Description, &versions, &apiRoot.MaxContentLength); err != nil {
			return apiRoot, err
		}
		apiRoot.Versions = strings.Split(versions, ",")
	}

	err = rows.Err()
	return apiRoot, err
}

func (s *sqliteDB) readAPIRoots(rows *sql.Rows) (interface{}, error) {
	var tas taxiiAPIRoots
	var err error

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return tas, err
		}
		tas.RootPaths = append(tas.RootPaths, path)
	}

	err = rows.Err()
	return tas, err
}

func (s *sqliteDB) readCollections(rows *sql.Rows) (interface{}, error) {
	tcs := taxiiCollections{}
	var err error

	for rows.Next() {
		var tc taxiiCollection
		var mediaTypes string

		if err := rows.Scan(&tc.ID, &tc.Title, &tc.Description, &tc.CanRead, &tc.CanWrite, &mediaTypes); err != nil {
			return tcs, err
		}
		tc.MediaTypes = strings.Split(mediaTypes, ",")
		tcs.Collections = append(tcs.Collections, tc)
	}

	err = rows.Err()
	return tcs, err
}

func (s *sqliteDB) readCollectionAccess(rows *sql.Rows) (interface{}, error) {
	var tcas []taxiiCollectionAccess
	var err error

	for rows.Next() {
		var tca taxiiCollectionAccess
		if err := rows.Scan(&tca.ID, &tca.CanRead, &tca.CanWrite); err != nil {
			return tca, err
		}
		tcas = append(tcas, tca)
	}

	err = rows.Err()
	return tcas, err
}

func (s *sqliteDB) readDiscovery(rows *sql.Rows) (interface{}, error) {
	td := taxiiDiscovery{}
	var apiRoots []string
	var err error

	for rows.Next() {
		var apiRoot string
		if err := rows.Scan(&td.Title, &td.Description, &td.Contact, &td.Default, &apiRoot); err != nil {
			return td, err
		}
		if apiRoot != "No API Roots defined" {
			apiRoots = append(apiRoots, apiRoot)
		}
	}

	err = rows.Err()
	td.APIRoots = apiRoots
	return td, err
}

func (s *sqliteDB) readManifest(rows *sql.Rows) (interface{}, error) {
	tm := taxiiManifest{}
	var err error

	for rows.Next() {
		tme := taxiiManifestEntry{}
		var versions string

		if err := rows.Scan(&tme.ID, &tme.DateAdded, &versions); err != nil {
			return tm, err
		}

		tme.Versions = strings.Split(string(versions), ",")
		tm.Objects = append(tm.Objects, tme)
	}

	err = rows.Err()
	return tm, err
}

func (s *sqliteDB) readRoutableCollections(rows *sql.Rows) (interface{}, error) {
	rs := routableCollections{}

	for rows.Next() {
		var collectionID taxiiID
		if err := rows.Scan(&collectionID); err != nil {
			return rs, err
		}
		rs.CollectionIDs = append(rs.CollectionIDs, collectionID)
	}

	err := rows.Err()
	return rs, err
}

type readFunction func(*sql.Rows) (interface{}, error)

func (s *sqliteDB) readRows(resource string, rows *sql.Rows) (result interface{}, err error) {
	defer rows.Close()

	resourceReader := map[string]readFunction{
		"routableCollections":   s.readRoutableCollections,
		"stixObject":            s.readStixObjects,
		"stixObjects":           s.readStixObjects,
		"taxiiAPIRoot":          s.readAPIRoot,
		"taxiiAPIRoots":         s.readAPIRoots,
		"taxiiCollection":       s.readCollections,
		"taxiiCollections":      s.readCollections,
		"taxiiCollectionAccess": s.readCollectionAccess,
		"taxiiDiscovery":        s.readDiscovery,
		"taxiiManifest":         s.readManifest,
		"taxiiUser":             s.readUser,
	}

	if resourceReader[resource] != nil {
		result, err = resourceReader[resource](rows)
		return
	}
	err = errors.New("Unknown resource name '" + resource)
	return
}

func (s *sqliteDB) readStixObjects(rows *sql.Rows) (interface{}, error) {
	sos := stixObjects{}
	var err error

	for rows.Next() {
		var object []byte
		if err := rows.Scan(&object); err != nil {
			return sos, err
		}
		sos.Objects = append(sos.Objects, object)
	}

	err = rows.Err()
	return sos, err
}

func (s *sqliteDB) readUser(rows *sql.Rows) (interface{}, error) {
	var valid bool
	var err error

	for rows.Next() {
		if err := rows.Scan(&valid); err != nil {
			return valid, err
		}
	}

	err = rows.Err()
	return valid, err
}

/* create methods */

func (s *sqliteDB) create(resource string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tq, err := s.parse("create", resource)
	if err != nil {
		errs <- err
		return
	}

	tx, stmt, err := batchWriteTx(s, tq.query, errs)
	if err != nil {
		return
	}
	defer stmt.Close()

	i := 0
	for item := range toWrite {
		args := item.([]interface{})

		_, err := stmt.Exec(args...)
		if err != nil {
			errs <- err
			log.WithFields(log.Fields{"args": args, "err": err}).Error("Failed to write")
			continue
		}

		i++
		if i >= maxWrites {
			tx.Commit() // on commit a statement is closed, create a new transaction for next batch
			tx, stmt, err = batchWriteTx(s, tq.query, errs)
			if err != nil {
				return
			}
		}
	}
	tx.Commit()
}

func batchWriteTx(s *sqliteDB, query string, errs chan error) (tx *sql.Tx, stmt *sql.Stmt, err error) {
	tx, err = s.db.Begin()
	if err != nil {
		errs <- err
		log.WithFields(log.Fields{"err": err}).Error("Failed to begin transaction")
		return
	}

	stmt, err = tx.Prepare(query)
	if err != nil {
		errs <- err
		log.WithFields(log.Fields{"err": err, "query": query}).Error("Failed to prepare query")
	}

	return
}
