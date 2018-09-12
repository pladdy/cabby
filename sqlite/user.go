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

const minPasswordLength = 8

// UserService implements a SQLite version of the servce
type UserService struct {
	DB        *sql.DB
	DataStore *DataStore
}

// CreateUser creates a user in the data store
func (s UserService) CreateUser(ctx context.Context, user cabby.User, password string) error {
	resource, action := "User", "create"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := validUserPasswordCombo(user, password)
	if err == nil {
		err = s.createUser(user, password)
	} else {
		log.WithFields(log.Fields{"error": err, "password": password, "user": user}).Error("Invalid user and/or password")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) createUser(u cabby.User, password string) error {
	sql := `insert into taxii_user (email, can_admin) values (?, ?)`
	args := []interface{}{u.Email, u.CanAdmin}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return err
	}

	sql = `insert into taxii_user_pass (email, pass) values (?, ?)`
	args = []interface{}{u.Email, hash(password)}

	err = s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// DeleteUser creates a user in the data store
func (s UserService) DeleteUser(ctx context.Context, user string) error {
	resource, action := "User", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteUser(user)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) deleteUser(user string) error {
	sql := `delete from taxii_user where email = ?`
	args := []interface{}{user}

	_, err := s.DB.Exec(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return err
	}

	sql = `delete from taxii_user_pass where email = ?`
	_, err = s.DB.Exec(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// UpdateUser creates a user in the data store
func (s UserService) UpdateUser(ctx context.Context, user cabby.User) error {
	resource, action := "User", "update"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := user.Validate()
	if err == nil {
		err = s.updateUser(user)
	} else {
		log.WithFields(log.Fields{"error": err, "user": user}).Error("Invalid user")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) updateUser(u cabby.User) error {
	sql := `update taxii_user set can_admin = ? where email = ?`
	args := []interface{}{u.CanAdmin, u.Email}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// User will read from the data store and populate the result with a resource
func (s UserService) User(ctx context.Context, user, password string) (cabby.User, error) {
	resource, action := "User", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.user(user, password)
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
	args := []interface{}{user, hash(password)}

	u := cabby.User{}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
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
func (s UserService) UserCollections(ctx context.Context, user string) (cabby.UserCollectionList, error) {
	resource, action := "UserCollectionList", "read"
	start := cabby.LogServiceStart(ctx, resource, action)
	result, err := s.userCollections(user)
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
	args := []interface{}{user}

	ucl := cabby.UserCollectionList{Email: user, CollectionAccessList: map[cabby.ID]cabby.CollectionAccess{}}

	rows, err := s.DB.Query(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
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

func validatePassword(password string) (err error) {
	if len(password) < minPasswordLength {
		return fmt.Errorf("Password length is too small, minimum length of characters is %d", minPasswordLength)
	}
	return
}

func validUserPasswordCombo(user cabby.User, password string) error {
	err := user.Validate()
	if err != nil {
		return err
	}
	return validatePassword(password)
}
