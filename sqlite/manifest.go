package sqlite

import (
	"database/sql"
	"strings"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// ManifestService implements a SQLite version of the ManifestService interface
type ManifestService struct {
	DB *sql.DB
}

// Manifest will read from the data store and return the resource
func (s ManifestService) Manifest(collectionID string, cr *cabby.Range) (cabby.Manifest, error) {
	resource, action := "Manifest", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.manifest(collectionID, cr)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s ManifestService) manifest(collectionID string, cr *cabby.Range) (cabby.Manifest, error) {
	sql := `with data as (
						select rowid, id, min(created) date_added, group_concat(modified) versions, 1 count
						-- media_types omitted...should that be in this table?
						from stix_objects_data
						where
							collection_id = ?
							/* $addedAfter
							$id
							$types
							$version */
						group by rowid, id
					)
					select id, date_added, versions, (select sum(count) from data) total
					from data`

	var args []interface{}

	if cr.Valid() {
		sql = WithPagination(sql)
		args = []interface{}{collectionID, (cr.Last - cr.First) + 1, cr.First}
	} else {
		args = []interface{}{collectionID}
	}

	m := cabby.Manifest{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		return m, err
	}

	for rows.Next() {
		me := cabby.ManifestEntry{}
		var versions string

		if err := rows.Scan(&me.ID, &me.DateAdded, &versions, &cr.Total); err != nil {
			return m, err
		}

		me.MediaTypes = []string{cabby.StixContentType}
		me.Versions = strings.Split(string(versions), ",")
		m.Objects = append(m.Objects, me)
	}

	err = rows.Err()
	return m, err
}
