package sqlite

import (
	"context"
	"database/sql"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

// VersionService implements a SQLite version of the VersionService interface
type VersionService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// Versions returns a given objects versions
func (s VersionService) Versions(ctx context.Context, cid, oid string, f cabby.Filter) (cabby.Versions, error) {
	resource, action := "Versions", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.versions(cid, oid, f)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s VersionService) versions(cid, oid string, f cabby.Filter) (cabby.Versions, error) {
	sql := `select modified version from objects_data
					where
						collection_id = ?
            and id = ?
						and $filter`

	args := []interface{}{cid, oid}
	sql, args = applyFiltering(sql, f, args)

	vs := cabby.Versions{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return vs, err
	}
	defer rows.Close()

	for rows.Next() {
		var version string

		if err := rows.Scan(&version); err != nil {
			return vs, err
		}

		ts, err := stones.TimestampFromString(version)
		if err != nil {
			return vs, err
		}

		vs.Versions = append(vs.Versions, ts.String())
	}

	return vs, rows.Err()
}
