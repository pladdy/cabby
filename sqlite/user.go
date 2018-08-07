package sqlite

import (
	"crypto/sha256"
	"database/sql"
	"fmt"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"

	cabby "github.com/pladdy/cabby2"
)

// UserService implements a SQLite version of the servce
type UserService struct {
	DB *sql.DB
}

// User will read from the data store and populate the result with a resource
func (s UserService) User(user, password string) (cabby.User, error) {
	resource, action := "user", "read"
	start := cabby.LogServiceStart(resource, action)

	sql := `select tu.email, tu.can_admin
          from
            taxii_user tu
            inner join taxii_user_pass tup
              on tu.email = tup.email
          where tu.email = ? and tup.pass = ?`

	u := cabby.User{}

	rows, err := s.DB.Query(sql, user, hash(password))
	if err != nil {
		return u, err
	}

	for rows.Next() {
		if err := rows.Scan(&u.Email, &u.CanAdmin); err != nil {
			return u, err
		}
	}

	err = rows.Err()
	cabby.LogServiceEnd(resource, action, start)
	return u, err
}

// Exists returns a bool indicating if a user is valid
func (s UserService) Exists(u cabby.User) bool {
	if u.Email == "" {
		return false
	}
	return true
}

/* helpers */

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
