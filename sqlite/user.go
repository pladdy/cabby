package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	cabby "github.com/pladdy/cabby2"
)

// UserService implements a SQLite version of the servce
type UserService struct {
	DB *sql.DB
}

// User will read from the data store and populate the result with a resource
func (s UserService) User(ctx context.Context, password string) (cabby.User, error) {
	resource, action := "User", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.user(cabby.TakeUser(ctx).Email, password)
	cabby.LogServiceEnd(ctx, resource, action, start)
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
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
		return u, err
	}
	defer rows.Close()

	for rows.Next() {
		if err := rows.Scan(&u.Email, &u.CanAdmin); err != nil {
			return u, err
		}
	}

	err = rows.Err()
	return u, err
}

// UserCollections will read from the data store and populate the result with a resource
func (s UserService) UserCollections(ctx context.Context) (cabby.UserCollectionList, error) {
	resource, action := "UserCollectionList", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.userCollections(cabby.TakeUser(ctx).Email)
	cabby.LogServiceEnd(ctx, resource, action, start)
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
		log.WithFields(log.Fields{"sql": sql, "error": err}).Error("error in sql")
		return ucl, err
	}
	defer rows.Close()

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
