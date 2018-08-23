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

// Exists returns a bool indicating if a user is valid
func (s UserService) Exists(u cabby.User) bool {
	if u.Email == "" {
		return false
	}
	return true
}

// User will read from the data store and populate the result with a resource
func (s UserService) User(user, password string) (cabby.User, error) {
	resource, action := "User", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.user(user, password)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s UserService) user(user, password string) (cabby.User, error) {
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
	return u, err
}

// UserCollections will read from the data store and populate the result with a resource
func (s UserService) UserCollections(user string) (cabby.UserCollectionList, error) {
	resource, action := "UserCollectionList", "read"
	start := cabby.LogServiceStart(resource, action)
	result, err := s.userCollections(user)
	cabby.LogServiceEnd(resource, action, start)
	return result, err
}

func (s UserService) userCollections(user string) (cabby.UserCollectionList, error) {
	sql := `select tuc.collection_id, tuc.can_read, tuc.can_write
					from
						taxii_user tu
						inner join taxii_user_collection tuc
							on tu.email = tuc.email
					where tu.email = ?`

	ucl := cabby.UserCollectionList{Email: user, CollectionAccessList: map[cabby.ID]cabby.CollectionAccess{}}

	rows, err := s.DB.Query(sql, user)
	if err != nil {
		return ucl, err
	}

	for rows.Next() {
		var ca cabby.CollectionAccess
		if err := rows.Scan(&ca.ID, &ca.CanRead, &ca.CanWrite); err != nil {
			return ucl, err
		}
		ucl.CollectionAccessList[ca.ID] = ca
	}

	return ucl, nil
}

/* helpers */

func hash(password string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(password)))
}
