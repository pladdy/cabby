package sqlite

import (
	"database/sql"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	cabby "github.com/pladdy/cabby2"
)

// StatusService implements a SQLite version of the StatusService interface
type StatusService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// CreateStatus will read from the data store and return the resource
func (s StatusService) CreateStatus(status cabby.Status) error {
	resource, action := "Status", "create"
	start := cabby.LogServiceStart(resource, action)
	err := s.createStatus(status)
	cabby.LogServiceEnd(resource, action, start)
	return err
}

func (s StatusService) createStatus(st cabby.Status) error {
	sql := `insert into taxii_status (id, status, total_count, success_count, failure_count, pending_count)
					values (?, ?, ?, ?, ?, ?)`

	return s.DataStore.write(sql, st.ID, st.Status, st.TotalCount, st.SuccessCount, st.FailureCount, st.PendingCount)
}

// Status will read from the data store and return the resource
func (s StatusService) Status(statusID string) (cabby.Status, error) {
	resource, action := "Status", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.status(statusID)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s StatusService) status(statusID string) (cabby.Status, error) {
	sql := `select id, status, total_count, success_count, pending_count, failure_count
					from taxii_status where id = ?`

	st := cabby.Status{}
	var err error

	rows, err := s.DB.Query(sql, statusID)
	if err != nil {
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
		return st, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(
			&st.ID, &st.Status, &st.TotalCount, &st.SuccessCount, &st.PendingCount, &st.FailureCount); err != nil {
			return st, err
		}
	}

	err = rows.Err()
	return st, err
}

// UpdateStatus will read from the data store and return the resource
func (s StatusService) UpdateStatus(status cabby.Status) error {
	resource, action := "Status", "update"
	start := cabby.LogServiceStart(resource, action)
	err := s.updateStatus(status)
	cabby.LogServiceEnd(resource, action, start)
	return err
}

func (s StatusService) updateStatus(st cabby.Status) error {
	sql := `update taxii_status
          set status = ?, total_count = ?, success_count = ?, failure_count = ?, pending_count = ?
          where id = ?`

	st.PendingCount = st.TotalCount - st.SuccessCount - st.FailureCount

	if st.PendingCount == 0 {
		st.SuccessCount = st.TotalCount - st.FailureCount
		st.Status = "complete"
	}

	return s.DataStore.write(sql, st.Status, st.TotalCount, st.SuccessCount, st.FailureCount, st.PendingCount, st.ID)
}
