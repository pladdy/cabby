package main

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"
	"time"

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
		"taxiiStatus": `insert into taxii_status (id, status, total_count, success_count, failure_count, pending_count)
		                values (?, ?, ?, ?, ?, ?)`,
		"taxiiUser": `insert into taxii_user (email, can_admin) values (?, ?)`,
		"taxiiUserCollection": `insert into taxii_user_collection (email, collection_id, can_read, can_write)
		                        values (?, ?, ?, ?)`,
		"taxiiUserPassword": `insert into taxii_user_pass (email, pass) values (?, ?)`,
	},
	"delete": map[string]string{
		"taxiiAPIRoot":        `delete from taxii_api_root where api_root_path = ?`,
		"taxiiCollection":     `delete from taxii_collection where id = ?`,
		"taxiiDiscovery":      `delete from taxii_discovery`,
		"taxiiUser":           `delete from taxii_user where email = ?`,
		"taxiiUserCollection": `delete from taxii_user_collection where email = ? and collection_id = ?`,
		"taxiiUserPassword":   `delete from taxii_user_pass where email = ?`,
	},
	"read": map[string]string{
		"routableCollections": `select id from taxii_collection where api_root_path = ?`,
		"stixObject": `select object
									 from stix_objects_data
									 where
									   collection_id = ?
										 and id = ?
										 $version`,
		"stixObjects": `with data as (
										  select rowid, object, 1 count
										  from stix_objects_data
											where
											  collection_id = ?
												$addedAfter
												$id
												$types
												$version
										)
										select object, (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
										from data
										$paginate`,
		"taxiiAPIRoot": `select api_root_path, title, description, versions, max_content_length
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
												 from data
												 $paginate`,
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
											  from stix_objects_data
											  where
												  collection_id = ?
													$addedAfter
													$id
													$types
													$version
											  group by rowid, id
											)
											select id, date_added, versions, (select min(rowid) from data), (select max(rowid) from data), (select sum(count) from data)
											from data
											$paginate`,
		"taxiiStatus": `select id, status, total_count, success_count, pending_count, failure_count
		                from taxii_status
										where id = ?`,
		"taxiiUser": `select tu.email, tu.can_admin
									from
									  taxii_user tu
									  inner join taxii_user_pass tup
									    on tu.email = tup.email
									where tu.email = ? and tup.pass = ?`,
		"taxiiUserCollection": `select email, collection_id, can_read, can_write
		                        from taxii_user_collection
														where email = ? and collection_id = ?`,
	},
	"update": map[string]string{
		"taxiiAPIRoot": `update taxii_api_root
		                 set title = ?, description = ?, versions = ?, max_content_length = ?
										 where api_root_path = ?`,
		"taxiiCollection": `update taxii_collection
		                    set id = ?, api_root_path = ?, title = ?, description = ?
		                    where id = ?`,
		"taxiiDiscovery": `update taxii_discovery
		                   set title = ?, description = ?, contact = ?, default_url = ?`,
		"taxiiStatus": `update taxii_status
		                set status = ?, total_count = ?, success_count = ?, failure_count = ?, pending_count = ?
		                where id = ?`,
		"taxiiUser": `update taxii_user set can_admin = ? where email = ?`,
		"taxiiUserCollection": `update taxii_user_collection
		                        set collection_id = ?, can_read = ?, can_write = ?
														where email = ?`,
		"taxiiUserPassword": `update taxii_user_pass set pass = ? where email = ?`,
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

func (s *sqliteDB) connect(path string) (err error) {
	// set foreign key pragma to true in connection
	// https://github.com/mattn/go-sqlite3#connection-string
	s.db, err = sql.Open(s.driver, path+"?_fk=true")
	if err != nil {
		log.Error(err)
	}
	return
}

func (s *sqliteDB) disconnect() {
	s.db.Close()
}

/* create methods */

func (s *sqliteDB) create(resource string, toWrite chan interface{}, errs chan error) {
	createFunction := withCreatorLogging(s.createResource)
	createFunction(resource, toWrite, errs)
}

func (s *sqliteDB) createResource(resource string, toWrite chan interface{}, errs chan error) {
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

/* delete */

func (s *sqliteDB) delete(resource string, args []interface{}) error {
	deleteFunction := withDeleterLogging(s.deleteResource)
	return deleteFunction(resource, args)
}

func (s *sqliteDB) deleteResource(resource string, args []interface{}) error {
	tq, err := newTaxiiQuery("delete", resource)
	if err != nil {
		return err
	}

	return executeQuery(s, tq.statement, args)
}

/* query methods */

func filterVersion(version string) string {
	t, err := time.Parse(time.RFC3339Nano, version)
	if err == nil {
		return `and (strftime('%s', strftime('%Y-%m-%d %H:%M:%f', modified))
	                        + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', modified))) * 1000 =
								(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', '` + t.Format(time.RFC3339Nano) + `'))
													+ strftime('%f', strftime('%Y-%m-%d %H:%M:%f', '` + t.Format(time.RFC3339Nano) + `'))) * 1000`
	}

	filter := " and version "

	switch version {
	case "all":
		filter = ""
	case "first":
		filter += "in ('first', 'only')"
	default:
		filter += "in ('last', 'only')"
	}
	return filter
}

// withFilter takes a query string, attempts to replace $<filter> place holders in it with taxiiFilters if they're
// set
func withFilter(query string, tf taxiiFilter) string {
	var addedAfter string
	var stixID string
	var stixTypes string
	var version string

	if len(tf.addedAfter) > 0 {
		addedAfter = `and (strftime('%s', strftime('%Y-%m-%d %H:%M:%f', created_at))
	                             + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', created_at))) * 1000 >
											(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', '` + tf.addedAfter + `'))
													 	   + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', '` + tf.addedAfter + `'))) * 1000`
	}
	query = strings.Replace(query, "$addedAfter", addedAfter, -1)

	if len(tf.stixID) > 0 {
		stixID = ` and id = '` + tf.stixID + "'"
	}
	query = strings.Replace(query, "$id", stixID, -1)

	if len(tf.stixTypes) > 0 {
		types := strings.Join(tf.stixTypes, "', '")
		stixTypes = ` and type in ('` + types + "')"
	}
	query = strings.Replace(query, "$types", stixTypes, -1)

	version = filterVersion(tf.version)
	query = strings.Replace(query, "$version", version, -1)

	return query
}

// withPagination for SQL uses limit/offset to paginate
// Warning:
//   this function requires a $paginate string placeholder in the query for pagination logic to be
//   interpolated in
func withPagination(query string, tr taxiiRange) string {
	var paginate string

	if tr.Valid() {
		limit := tr.last - tr.first
		paginate = fmt.Sprintf("limit %v offset %v", limit+1, tr.first)
	}

	return strings.Replace(query, "$paginate", paginate, -1)
}

/* read methods */

func (s *sqliteDB) read(resource string, args []interface{}, tf ...taxiiFilter) (result taxiiResult, err error) {
	readFunction := withReaderLogging(s.readResource)
	return readFunction(resource, args, tf...)
}

func (s *sqliteDB) readResource(resource string, args []interface{}, tf ...taxiiFilter) (result taxiiResult, err error) {
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
		log.WithFields(log.Fields{"statement": tq.statement, "err": err}).Error("Failed to query")
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
		if err := rows.Scan(&apiRoot.Path, &apiRoot.Title, &apiRoot.Description, &versions, &apiRoot.MaxContentLength); err != nil {
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
		"taxiiStatus":           s.readStatus,
		"taxiiUser":             s.readUser,
		"taxiiUserCollection":   s.readUserCollection,
	}

	if resourceReader[tq.resource] != nil {
		result, err = resourceReader[tq.resource](rows)
		return
	}
	err = errors.New("Unknown resource name: " + tq.resource)
	return
}

func (s *sqliteDB) readStatus(rows *sql.Rows) (taxiiResult, error) {
	var st taxiiStatus
	var err error

	for rows.Next() {
		if err := rows.Scan(&st.ID, &st.Status, &st.TotalCount, &st.SuccessCount, &st.FailureCount, &st.PendingCount); err != nil {
			return taxiiResult{data: st}, err
		}
	}

	err = rows.Err()
	return taxiiResult{data: st}, err
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
	tu := taxiiUser{}
	var err error

	for rows.Next() {
		if err := rows.Scan(&tu.Email, &tu.CanAdmin); err != nil {
			return taxiiResult{data: tu}, err
		}
	}

	err = rows.Err()
	return taxiiResult{data: tu}, err
}

func (s *sqliteDB) readUserCollection(rows *sql.Rows) (taxiiResult, error) {
	tuc := taxiiUserCollection{}
	tca := tuc.taxiiCollectionAccess
	var err error

	for rows.Next() {
		if err := rows.Scan(&tuc.Email, &tca.ID, &tca.CanRead, &tca.CanWrite); err != nil {
			return taxiiResult{data: tuc}, err
		}
	}
	tuc.taxiiCollectionAccess = tca

	err = rows.Err()
	return taxiiResult{data: tuc}, err
}

/* update */

func (s *sqliteDB) update(resource string, args []interface{}) error {
	updateFunction := withUpdaterLogging(s.updateResource)
	return updateFunction(resource, args)
}

func (s *sqliteDB) updateResource(resource string, args []interface{}) error {
	tq, err := newTaxiiQuery("update", resource)
	if err != nil {
		return err
	}

	return executeQuery(s, tq.statement, args)
}

/* helpers */

func executeQuery(s *sqliteDB, q string, args []interface{}) error {
	_, err := s.db.Exec(q, args...)
	if err != nil {
		log.WithFields(log.Fields{"statement": q, "args": args, "err": err}).Error("Failed to execute")
		return fmt.Errorf("%v in statement: %v", err, q)
	}
	return err
}
