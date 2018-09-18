package sqlite

import (
	"context"
	"database/sql"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	cabby "github.com/pladdy/cabby2"
)

// DiscoveryService implements a SQLite version of the DiscoveryService interface
type DiscoveryService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// CreateDiscovery creates a user in the data store
func (s DiscoveryService) CreateDiscovery(ctx context.Context, d cabby.Discovery) error {
	resource, action := "Discovery", "create"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := d.Validate()
	if err == nil {
		err = s.createDiscovery(d)
	} else {
		log.WithFields(log.Fields{"discovery": d, "error": err}).Error("Invalid Discovery")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s DiscoveryService) createDiscovery(d cabby.Discovery) error {
	sql := `insert into taxii_discovery (title, description, contact, default_url) values (?, ?, ?, ?)`
	args := []interface{}{d.Title, d.Description, d.Contact, d.Default}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// DeleteDiscovery creates a user in the data store
func (s DiscoveryService) DeleteDiscovery(ctx context.Context) error {
	resource, action := "Discovery", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteDiscovery()
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s DiscoveryService) deleteDiscovery() error {
	sql := `delete from taxii_discovery`
	args := []interface{}{}

	_, err := s.DB.Exec(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// Discovery will read from the data store and return the resource
func (s DiscoveryService) Discovery(ctx context.Context) (cabby.Discovery, error) {
	resource, action := "Discovery", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.discovery()
	cabby.LogServiceEnd(ctx, resource, action, start)
	return result, err
}

func (s DiscoveryService) discovery() (cabby.Discovery, error) {
	sql := `select td.title, td.description, td.contact, td.default_url,
						 case
							 when tar.api_root_path is null then 'No API Roots defined' else tar.api_root_path
						 end discovery_path
					 from
						 taxii_discovery td
						 left join taxii_api_root tar
							 on td.id = tar.discovery_id`
	args := []interface{}{}

	d := cabby.Discovery{}
	var apiRoots []string

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return d, err
	}
	defer rows.Close()

	for rows.Next() {
		var apiRoot string
		if err := rows.Scan(&d.Title, &d.Description, &d.Contact, &d.Default, &apiRoot); err != nil {
			return d, err
		}
		if apiRoot != "No API Roots defined" {
			apiRoots = append(apiRoots, apiRoot)
		}
	}

	err = rows.Err()
	d.APIRoots = apiRoots
	return d, err
}

// UpdateDiscovery creates a user in the data store
func (s DiscoveryService) UpdateDiscovery(ctx context.Context, d cabby.Discovery) error {
	resource, action := "Discovery", "update"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := d.Validate()
	if err == nil {
		err = s.updateDiscovery(d)
	} else {
		log.WithFields(log.Fields{"discovery": d, "error": err}).Error("Invalid Discovery")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s DiscoveryService) updateDiscovery(d cabby.Discovery) error {
	sql := `update taxii_discovery
					set title = ?, description = ?, contact = ?, default_url = ?`
	args := []interface{}{d.Title, d.Description, d.Contact, d.Default}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}
