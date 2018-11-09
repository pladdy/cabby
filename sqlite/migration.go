package sqlite

import (
	"database/sql"
	"sort"

	"github.com/pladdy/cabby/sqlite/migrations"
	log "github.com/sirupsen/logrus"
)

// add migrations here to the below far
// each struct has the version number associated to it and it's functions for migration up and down
var migrationsToSetup = []migrationList{
	migrationList{1, migrations.Up1, migrations.Down1}}

type migrationList struct {
	version int
	upFn    migrationFn
	downFn  migrationFn
}

type migrationFn func() string

// MigrationService implements a SQLite version of the MigrationService interface
type MigrationService struct {
	DataStore *DataStore
	DB        *sql.DB
	Versions  []int
	downs     map[int]migrationFn
	ups       map[int]migrationFn
}

// NewMigrationService returns a service for migrations
func NewMigrationService() MigrationService {
	ms := MigrationService{}
	ms.downs = make(map[int]migrationFn)
	ms.ups = make(map[int]migrationFn)
	ms.setup()
	return ms
}

// CurrentVersion returns current version of the database
func (s MigrationService) CurrentVersion() (version int, err error) {
	if !s.schemaIsVersioned() {
		return 0, nil
	}

	sql := `select max(version) from schema_version`

	rows, err := s.DB.Query(sql)
	if err != nil {
		logSQLError(sql, []interface{}{}, err)
		return
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&version); err != nil {
			return version, err
		}
	}

	err = rows.Err()
	return
}

// Up migrates the database up if necessary
func (s MigrationService) Up() (err error) {
	cv, _ := s.CurrentVersion()

	for i := cv; i < len(s.Versions); i++ {
		version := s.Versions[i]
		migration := s.ups[version]

		log.WithFields(log.Fields{"migration": version}).Info("Running migration")

		_, err = s.DB.Exec(migration())
		if err != nil {
			break
		}
	}
	return
}

func (s MigrationService) registerMigration(version int, up, down migrationFn) {
	s.ups[version] = up
	s.downs[version] = down
}

func (s MigrationService) schemaIsVersioned() bool {
	sql := `select name from sqlite_master where type = 'table' and name = 'schema_version'`

	rows, err := s.DB.Query(sql)
	if err != nil {
		logSQLError(sql, []interface{}{}, err)
		return false
	}
	defer rows.Close()

	var table string
	for rows.Next() {
		if err := rows.Scan(&table); err != nil {
			return false
		}
	}
	err = rows.Err()

	if table == "" || err != nil {
		log.Warn("schema is not versioned...")
		return false
	}
	return true
}

func (s *MigrationService) setup() {
	for _, list := range migrationsToSetup {
		s.registerMigration(list.version, list.upFn, list.downFn)
	}
	s.Versions = sortVersions(s.ups)
}

/* helpers */

func sortVersions(m map[int]migrationFn) (versions []int) {
	for k := range m {
		versions = append(versions, k)
	}
	sort.Ints(versions)
	return
}
