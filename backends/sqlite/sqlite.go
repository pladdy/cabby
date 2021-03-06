package sqlite

import (
	"database/sql"
	"errors"
	"regexp"
	"strings"
	"time"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
)

const (
	maxWritesPerBatch = 500
)

// DataStore represents a SQLite database
type DataStore struct {
	DB   *sql.DB
	Path string
}

// NewDataStore returns a sqliteDB
func NewDataStore(path string) (*DataStore, error) {
	s := DataStore{Path: path}
	if s.Path == "" {
		return &s, errors.New("No database location specfied in config")
	}

	err := s.Open()
	return &s, err
}

// APIRootService returns a service for api root resources
func (s *DataStore) APIRootService() cabby.APIRootService {
	return APIRootService{DB: s.DB, DataStore: s}
}

// Close connection to datastore
func (s *DataStore) Close() {
	err := s.DB.Close()
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Error("Failed to close the database connection")
	}
}

// CollectionService returns a service for collection resources
func (s *DataStore) CollectionService() cabby.CollectionService {
	return CollectionService{DB: s.DB, DataStore: s}
}

// DiscoveryService returns a service for discovery resources
func (s *DataStore) DiscoveryService() cabby.DiscoveryService {
	return DiscoveryService{DB: s.DB, DataStore: s}
}

// ManifestService returns a service for object resources
func (s *DataStore) ManifestService() cabby.ManifestService {
	return ManifestService{DB: s.DB}
}

// MigrationService returns service to database mgirations
func (s *DataStore) MigrationService() cabby.MigrationService {
	ms := NewMigrationService()
	ms.DB = s.DB
	return ms
}

// ObjectService returns a service for object resources
func (s *DataStore) ObjectService() cabby.ObjectService {
	return ObjectService{DB: s.DB, DataStore: s}
}

// Open connection to datastore
func (s *DataStore) Open() (err error) {
	// set foreign key pragma to true in connection: https://github.com/mattn/go-sqlite3#connection-string
	s.DB, err = sql.Open("sqlite3", s.Path+"?_fk=true")
	if err != nil {
		log.Error(err)
	}
	return
}

// StatusService returns service for status resources
func (s *DataStore) StatusService() cabby.StatusService {
	return StatusService{DB: s.DB, DataStore: s}
}

// UserService returns a service for user resources
func (s *DataStore) UserService() cabby.UserService {
	return UserService{DB: s.DB, DataStore: s}
}

// VersionsService returns a service for versions
func (s *DataStore) VersionsService() cabby.VersionsService {
	return VersionsService{DB: s.DB, DataStore: s}
}

/* writer methods */

func (s *DataStore) batchWrite(query string, toWrite chan interface{}, errs chan error) {
	defer close(errs)

	tx, stmt, err := s.writeOperation(query)
	if err != nil {
		errs <- err
		return
	}
	defer stmt.Close()

	i := 0
	for item := range toWrite {
		args := item.([]interface{})

		err := s.execute(stmt, args...)
		if err != nil {
			log.WithFields(log.Fields{"sql": query, "error": err}).Error("Error after call to 'execute'")
			errs <- err
			continue
		}

		i++
		if i >= maxWritesPerBatch {
			err := tx.Commit() // on commit a statement is closed, create a new transaction for next batch
			if err != nil {
				log.WithFields(log.Fields{"sql": query, "error": err}).Error("Error after call to 'Commit'")
				errs <- err
				return
			}

			tx, stmt, err = s.writeOperation(query)
			if err != nil {
				log.WithFields(log.Fields{"sql": query, "error": err}).Error("Error after call to 'writeOperation'")
				errs <- err
				return
			}
			i = 0
		}
	}
	errs <- tx.Commit()
}

func (s *DataStore) execute(stmt *sql.Stmt, args ...interface{}) error {
	_, err := stmt.Exec(args...)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "args": args}).Error("Failed to execute")
	}
	return err
}

// helper function used by tests to create resources
func (s *DataStore) write(query string, args ...interface{}) error {
	tx, stmt, err := s.writeOperation(query)
	if err != nil {
		log.WithFields(log.Fields{"sql": query, "error": err}).Error("error in sql")
		return err
	}
	defer tx.Commit()
	defer stmt.Close()

	return s.execute(stmt, args...)
}

func (s *DataStore) writeOperation(query string) (tx *sql.Tx, stmt *sql.Stmt, err error) {
	tx, err = s.DB.Begin()
	if err != nil {
		log.WithFields(log.Fields{"err": err}).Error("Failed to begin transaction")
		return
	}

	stmt, err = tx.Prepare(query)
	if err != nil {
		log.WithFields(log.Fields{"err": err, "sql": query}).Error("Failed to prepare query")
	}
	return
}

// Filter implementation for SQLite
type Filter struct {
	cabby.Filter
}

// QueryString will convert a filter into a query string for a service query
func (f *Filter) QueryString() (q string, args []interface{}) {
	var filters []string

	if !f.AddedAfter.IsZero() {
		filter, newArgs := filterAddedAfter(f.AddedAfter.String())
		if filter != "" {
			filters = append(filters, filter)
			args = append(args, newArgs...)
		}
	}

	if len(f.Versions) > 0 {
		filter, newArgs := filterVersion(f.Versions)
		if filter != "" {
			filters = append(filters, filter)
			args = append(args, newArgs...)
		}
	}

	for field, raws := range filterMapFieldToRawStrings(f) {
		if len(raws) > 0 {
			filter, newArgs := filterCreator(raws, field)
			filters = append(filters, filter)
			args = append(args, newArgs...)
		}
	}

	return strings.Join(filters, " and "), args
}

/* filtering helpers */

func applyFiltering(sql string, f cabby.Filter, args []interface{}) (string, []interface{}) {
	filter := Filter{f}
	qs, filterArgs := filter.QueryString()

	if len(qs) > 0 {
		sql = strings.Replace(sql, "$filter", qs, -1)
	} else {
		sql = strings.Replace(sql, "$filter", "", -1)
	}

	sql = filterRemoveTrailingAnd(sql)
	args = append(args, filterArgs...)
	return sql, args
}

func filterAddedAfter(addedAfter string) (string, []interface{}) {
	filter := `(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', created_at))
					     + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', created_at))) * 1000 >
				   	 (strftime('%s', strftime('%Y-%m-%d %H:%M:%f', ?))
				  		 + strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000`
	return filter, []interface{}{addedAfter, addedAfter}
}

func filterCreator(raw, field string) (filter string, args []interface{}) {
	raws := strings.Split(raw, ",")
	filter = "("
	var ors []string

	for _, t := range raws {
		ors = append(ors, field+" = ?")
		args = append(args, t)
	}

	return filter + strings.Join(ors, " or ") + ")", args
}

func filterMapFieldToRawStrings(f *Filter) map[string]string {
	return map[string]string{
		"id":   f.IDs,
		"type": f.Types}
}

func filterRemoveTrailingAnd(sql string) string {
	lines := strings.Split(sql, "\n")
	re := regexp.MustCompile(`and\s*$`)

	for i := 0; i < len(lines); i++ {
		lines[i] = re.ReplaceAllString(lines[i], "")
	}

	return strings.Join(lines, "\n")
}

func filterVersion(rawVersion string) (filter string, args []interface{}) {
	versionFilterSQL := `(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', modified))
		+ strftime('%f', strftime('%Y-%m-%d %H:%M:%f', modified))) * 1000 =
	(strftime('%s', strftime('%Y-%m-%d %H:%M:%f', ?))
		+ strftime('%f', strftime('%Y-%m-%d %H:%M:%f', ?))) * 1000`

	versions := strings.Split(rawVersion, ",")
	var ors []string

	for _, v := range versions {
		t, err := time.Parse(time.RFC3339Nano, v)
		if err == nil {
			ors = append(ors, versionFilterSQL)
			args = append(args, []interface{}{t.Format(time.RFC3339Nano), t.Format(time.RFC3339Nano)}...)
		} else {
			switch v {
			case "all":
				ors = append(ors, "1 = 1")
			case "first":
				ors = append(ors, "version in ('first', 'only')")
			default:
				ors = append(ors, "version in ('last', 'only')")
			}
		}
	}

	return "(" + strings.Join(ors, " or ") + ")", args
}

func logSQLError(sql string, args []interface{}, err error) {
	log.WithFields(log.Fields{"error": err, "sql": sql, "args": args}).Error("Error in sql")
}

// Page implementation for SQLite
type Page struct {
	*cabby.Page
}

// QueryString returns sql for paginating a range of data
func (p *Page) QueryString() (q string, args []interface{}) {
	if p.Valid() {
		q = "limit ?"
		args = []interface{}{(p.Limit)}
	}
	return
}

/* pagination helpers */

func applyPaging(sql string, cp *cabby.Page, args []interface{}) (newSQL string, newArgs []interface{}) {
	p := Page{cp}
	qs, pageArgs := p.QueryString()

	if len(pageArgs) > 0 {
		sql = strings.Replace(sql, "$paginate", qs, -1)
	} else {
		sql = strings.Replace(sql, "$paginate", "", -1)
	}

	args = append(args, pageArgs...)
	return sql, args
}
