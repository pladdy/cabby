package sqlite

import (
	"database/sql"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// DiscoveryService implements a SQLite version of the servce
type DiscoveryService struct {
	DB *sql.DB
}

// Read will read from the data store and populate the discovery resource
func (s DiscoveryService) Read() (cabby.Result, error) {
	return cabby.WithReadLogging("discovery", s.read)()
}

func (s *DiscoveryService) read() (cabby.Result, error) {
	sql := `select td.title, td.description, td.contact, td.default_url,
						 case
							 when tar.api_root_path is null then 'No API Roots defined' else tar.api_root_path
						 end api_root_path
					 from
						 taxii_discovery td
						 left join taxii_api_root tar
							 on td.id = tar.discovery_id`

	d := cabby.Discovery{}
	var apiRoots []string

	rows, err := s.DB.Query(sql)
	if err != nil {
		return cabby.Result{Data: d}, err
	}

	for rows.Next() {
		var apiRoot string
		if err := rows.Scan(&d.Title, &d.Description, &d.Contact, &d.Default, &apiRoot); err != nil {
			return cabby.Result{Data: d}, err
		}
		if apiRoot != "No API Roots defined" {
			apiRoots = append(apiRoots, apiRoot)
		}
	}

	err = rows.Err()
	d.APIRoots = apiRoots
	return cabby.Result{Data: d}, err
}
