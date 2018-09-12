package sqlite

import (
	"context"
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
func (s ManifestService) Manifest(ctx context.Context, collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
	resource, action := "Manifest", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.manifest(collectionID, cr, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s ManifestService) manifest(collectionID string, cr *cabby.Range, f cabby.Filter) (cabby.Manifest, error) {
	sql := `with data as (
						select rowid, id, min(created) date_added, group_concat(modified) versions, 1 count
						-- media_types omitted...should that be in this table?
						from stix_objects_data
						where
							collection_id = ?
							and $filter
						group by rowid, id
					)
					select id, date_added, versions, (select sum(count) from data) total
					from data
					$paginate`

	args := []interface{}{collectionID}

	sql, args = applyFiltering(sql, f, args)
	sql, args = applyPaging(sql, cr, args)

	m := cabby.Manifest{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return m, err
	}
	defer rows.Close()

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
