package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"
)

type readFunction func(*sql.Rows) (taxiiResult, error)

type sqliteDB struct {
	db        *sql.DB
	dbName    string
	extension string
	path      string
	driver    string
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
		"stixObject": `select object
                   from stix_objects
									 where
									   collection_id = ?
										 and id = ?
										 $filter`,
		"stixObjects": `with data as (
										  select rowid, object, 1 count
										  from stix_objects
											where
											  collection_id = ?
												$filter
										)
										select object, (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
										from data`,
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
		"taxiiCollections": `with data as (
													 select rowid, id, title, description, can_read, can_write, media_types, 1 count
													 from (
														 select c.rowid, c.id, c.title, c.description, uc.can_read, uc.can_write, c.media_types
														 from
															 taxii_collection c
															 inner join taxii_user_collection uc
																 on c.id = uc.collection_id
														 where
															 uc.email = ?
															 and (uc.can_read = 1 or uc.can_write = 1)
													 )
												 )
												 select
												   id, title, description, can_read, can_write, media_types,
												   (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
												 from data`,
		"taxiiDiscovery": `select td.title, td.description, td.contact, td.default_url,
											   case
											     when tar.api_root_path is null then 'No API Roots defined' else tar.api_root_path
											   end api_root_path
											 from
											   taxii_discovery td
											   left join taxii_api_root tar
											     on td.id = tar.discovery_id`,
		"taxiiManifest": `with data as (
			                  select rowid, id, min(created) date_added, group_concat(modified) versions, 1 count -- media_types omitted for now
											  from stix_objects
											  where
												  collection_id = ?
													$filter
											  group by id
											)
											select id, date_added, versions, (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
											from data`,
		"taxiiUser": `select 1
									from
									  taxii_user tu
									  inner join taxii_user_pass tup
									    on tu.email = tup.email
									where tu.email = ? and tup.pass = ?`,
	},
}

func newTaxiiQuery(command, resource string) (taxiiQuery, error) {
	var err error
	statement := statements[command][resource]

	if statement == "" {
		return taxiiQuery{}, fmt.Errorf("invalid command %v and resource %v", command, resource)
	}
	return taxiiQuery{resource: resource, statement: statement}, err
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

/* query methods */

// withFilter takes a query string, attempts to replace the $filter place holder in it with taxiiFilters if they're
// set
func withFilter(query string, tf taxiiFilter) string {
	var filter string

	if len(tf.addedAfter) > 0 {
		filter += fmt.Sprintf("and created_at > %v", tf.addedAfter)
	}

	return strings.Replace(query, "$filter", filter, -1)
}

// withPagination for SQL has to increment the first and last fields of the taxiiRange because SQL rowids
// start at index 1, not 0.
// Warning:
//   this function operates off of a huge assumption, which is 'from data' is the from clause
//   'from data' is short for 'from a subquery called data that has rowid in it...'; this is gross
func withPagination(query string, tr taxiiRange) string {
	if !tr.Valid() {
		return query
	}

	return strings.Replace(query,
		"from data",
		fmt.Sprintf("from data where rowid between %v and %v", tr.first+1, tr.last+1),
		-1)
}

/* read methods */

func (s *sqliteDB) read(resource string, args []interface{}, tf ...taxiiFilter) (result taxiiResult, err error) {
	tq, err := newTaxiiQuery("read", resource)
	if err != nil {
		return result, err
	}

	if len(tf) > 0 {
		tq.statement = withPagination(tq.statement, tf[0].pagination)
		tq.statement = withFilter(tq.statement, tf[0])
	}

	rows, err := s.db.Query(tq.statement, args...)
	if err != nil {
		return result, fmt.Errorf("%v in statement: %v", err, tq.statement)
	}

	result, err = s.readRows(tq, rows)

	if len(tf) > 0 {
		result.withPagination(tf[0].pagination)
	}
	return result, err
}

/* read helpers */

func (s *sqliteDB) readAPIRoot(rows *sql.Rows) (taxiiResult, error) {
	var apiRoot taxiiAPIRoot
	var err error

	for rows.Next() {
		var versions string
		if err := rows.Scan(&apiRoot.Title, &apiRoot.Description, &versions, &apiRoot.MaxContentLength); err != nil {
			return taxiiResult{data: apiRoot}, err
		}
		apiRoot.Versions = strings.Split(versions, ",")
	}

	err = rows.Err()
	return taxiiResult{data: apiRoot}, err
}

func (s *sqliteDB) readAPIRoots(rows *sql.Rows) (taxiiResult, error) {
	var tas taxiiAPIRoots
	var err error

	for rows.Next() {
		var path string
		if err := rows.Scan(&path); err != nil {
			return taxiiResult{data: tas}, err
		}
		tas.RootPaths = append(tas.RootPaths, path)
	}

	err = rows.Err()
	return taxiiResult{data: tas}, err
}

func (s *sqliteDB) readCollection(rows *sql.Rows) (taxiiResult, error) {
	tcs := taxiiCollections{}
	tr := taxiiResult{}
	var err error

	for rows.Next() {
		var tc taxiiCollection
		var mediaTypes string

		if err := rows.Scan(&tc.ID, &tc.Title, &tc.Description, &tc.CanRead, &tc.CanWrite, &mediaTypes); err != nil {
			tr.data = tcs
			return tr, err
		}
		tc.MediaTypes = strings.Split(mediaTypes, ",")
		tcs.Collections = append(tcs.Collections, tc)
	}

	tr.data = tcs
	err = rows.Err()
	return tr, err
}

func (s *sqliteDB) readCollections(rows *sql.Rows) (taxiiResult, error) {
	tcs := taxiiCollections{}
	tr := taxiiResult{}
	var err error

	for rows.Next() {
		var tc taxiiCollection
		var mediaTypes string

		if err := rows.Scan(&tc.ID, &tc.Title, &tc.Description, &tc.CanRead, &tc.CanWrite, &mediaTypes,
			&tr.itemStart, &tr.itemEnd, &tr.items); err != nil {
			tr.data = tcs
			return tr, err
		}
		tc.MediaTypes = strings.Split(mediaTypes, ",")
		tcs.Collections = append(tcs.Collections, tc)
	}

	tr.data = tcs
	err = rows.Err()
	return tr, err
}

func (s *sqliteDB) readCollectionAccess(rows *sql.Rows) (taxiiResult, error) {
	var tcas []taxiiCollectionAccess
	var err error

	for rows.Next() {
		var tca taxiiCollectionAccess
		if err := rows.Scan(&tca.ID, &tca.CanRead, &tca.CanWrite); err != nil {
			return taxiiResult{data: tcas}, err
		}
		tcas = append(tcas, tca)
	}

	err = rows.Err()
	return taxiiResult{data: tcas}, err
}

func (s *sqliteDB) readDiscovery(rows *sql.Rows) (taxiiResult, error) {
	td := taxiiDiscovery{}
	var apiRoots []string
	var err error

	for rows.Next() {
		var apiRoot string
		if err := rows.Scan(&td.Title, &td.Description, &td.Contact, &td.Default, &apiRoot); err != nil {
			return taxiiResult{data: td}, err
		}
		if apiRoot != "No API Roots defined" {
			apiRoots = append(apiRoots, apiRoot)
		}
	}

	err = rows.Err()
	td.APIRoots = apiRoots
	return taxiiResult{data: td}, err
}

func (s *sqliteDB) readManifest(rows *sql.Rows) (taxiiResult, error) {
	tm := taxiiManifest{}
	tr := taxiiResult{}
	var err error

	for rows.Next() {
		tme := taxiiManifestEntry{}
		var versions string

		if err := rows.Scan(&tme.ID, &tme.DateAdded, &versions, &tr.itemStart, &tr.itemEnd, &tr.items); err != nil {
			tr.data = tm
			return tr, err
		}

		tme.Versions = strings.Split(string(versions), ",")
		tm.Objects = append(tm.Objects, tme)
	}

	tr.data = tm
	err = rows.Err()
	return tr, err
}

func (s *sqliteDB) readRoutableCollections(rows *sql.Rows) (taxiiResult, error) {
	rs := routableCollections{}

	for rows.Next() {
		var collectionID taxiiID
		if err := rows.Scan(&collectionID); err != nil {
			return taxiiResult{data: rs}, err
		}
		rs.CollectionIDs = append(rs.CollectionIDs, collectionID)
	}

	err := rows.Err()
	return taxiiResult{data: rs}, err
}

func (s *sqliteDB) readRows(tq taxiiQuery, rows *sql.Rows) (result taxiiResult, err error) {
	defer rows.Close()
	result.query = tq

	resourceReader := map[string]readFunction{
		"routableCollections":   s.readRoutableCollections,
		"stixObject":            s.readStixObject,
		"stixObjects":           s.readStixObjects,
		"taxiiAPIRoot":          s.readAPIRoot,
		"taxiiAPIRoots":         s.readAPIRoots,
		"taxiiCollection":       s.readCollection,
		"taxiiCollections":      s.readCollections,
		"taxiiCollectionAccess": s.readCollectionAccess,
		"taxiiDiscovery":        s.readDiscovery,
		"taxiiManifest":         s.readManifest,
		"taxiiUser":             s.readUser,
	}

	if resourceReader[tq.resource] != nil {
		result, err = resourceReader[tq.resource](rows)
		return
	}
	err = errors.New("Unknown resource name '" + tq.resource)
	return
}

func (s *sqliteDB) readStixObject(rows *sql.Rows) (taxiiResult, error) {
	sos := stixObjects{}
	var err error

	for rows.Next() {
		var object []byte
		if err := rows.Scan(&object); err != nil {
			return taxiiResult{data: sos}, err
		}
		sos.Objects = append(sos.Objects, object)
	}

	err = rows.Err()
	return taxiiResult{data: sos}, err
}

func (s *sqliteDB) readStixObjects(rows *sql.Rows) (taxiiResult, error) {
	sos := stixObjects{}
	tr := taxiiResult{}
	var err error

	for rows.Next() {
		var object []byte
		if err := rows.Scan(&object, &tr.itemStart, &tr.itemEnd, &tr.items); err != nil {
			return taxiiResult{data: sos}, err
		}
		sos.Objects = append(sos.Objects, object)
	}

	err = rows.Err()
	tr.data = sos
	return tr, err
}

func (s *sqliteDB) readUser(rows *sql.Rows) (taxiiResult, error) {
	var valid bool
	var err error

	for rows.Next() {
		if err := rows.Scan(&valid); err != nil {
			return taxiiResult{data: valid}, err
		}
	}

	err = rows.Err()
	return taxiiResult{data: valid}, err
}

/* create methods */

func (s *sqliteDB) create(resource string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tq, err := newTaxiiQuery("create", resource)
	if err != nil {
		errs <- err
		return
	}

	tx, stmt, err := batchWriteTx(s, tq.statement, errs)
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
			tx, stmt, err = batchWriteTx(s, tq.statement, errs)
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
