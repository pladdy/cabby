package sqlite

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"fmt"

	// import sqlite dependency
	_ "github.com/mattn/go-sqlite3"
	log "github.com/sirupsen/logrus"

	"github.com/pladdy/cabby"
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

	err := validateUserPasswordCombo(user, password)
	if err == nil {
		err = s.createUser(user, password)
	} else {
		log.WithFields(log.Fields{"error": err, "password": password, "user": user}).Error("Invalid user and/or password")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) createUser(u cabby.User, password string) error {
	sql := `insert into user (email, can_admin) values (?, ?)`
	args := []interface{}{u.Email, u.CanAdmin}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return err
	}

	sql = `insert into user_pass (email, pass) values (?, ?)`
	args = []interface{}{u.Email, hash(password)}

	err = s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// CreateUserCollection creates an association of a user to a collection
func (s UserService) CreateUserCollection(ctx context.Context, user string, ca cabby.CollectionAccess) error {
	resource, action := "UserCollection", "create"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := validateUserCollection(user, ca)
	if err == nil {
		err = s.createUserCollection(user, ca)
	} else {
		log.WithFields(log.Fields{"error": err, "collection_access": ca, "user": user}).Error("Invalid user and/or collection")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) createUserCollection(user string, ca cabby.CollectionAccess) error {
	sql := `insert into user_collection (email, collection_id, can_read, can_write)
				  values (?, ?, ?, ?)`
	args := []interface{}{user, ca.ID.String(), ca.CanRead, ca.CanWrite}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// DeleteUserCollection deletes a collection from a user
func (s UserService) DeleteUserCollection(ctx context.Context, user, id string) error {
	resource, action := "UserCollection", "delete"
	start := cabby.LogServiceStart(ctx, resource, action)
	err := s.deleteUserCollection(user, id)
	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) deleteUserCollection(user, id string) error {
	sql := `delete from user_collection where email = ? and collection_id = ?`
	args := []interface{}{user, id}

	_, err := s.DB.Exec(sql, args...)
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
	sql := `delete from user where email = ?`
	args := []interface{}{user}

	_, err := s.DB.Exec(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
		return err
	}

	sql = `delete from user_pass where email = ?`
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
	sql := `update user set can_admin = ? where email = ?`
	args := []interface{}{u.CanAdmin, u.Email}

	err := s.DataStore.write(sql, args...)
	if err != nil {
		logSQLError(sql, args, err)
	}
	return err
}

// UpdateUserCollection update a users access to a specfific collection
func (s UserService) UpdateUserCollection(ctx context.Context, user string, ca cabby.CollectionAccess) error {
	resource, action := "UserCollection", "update"
	start := cabby.LogServiceStart(ctx, resource, action)

	err := validateUserCollection(user, ca)
	if err == nil {
		err = s.updateUserCollection(user, ca)
	} else {
		log.WithFields(log.Fields{"error": err, "collection_access": ca, "user": user}).Error("Invalid user and/or collection")
	}

	cabby.LogServiceEnd(ctx, resource, action, start)
	return err
}

func (s UserService) updateUserCollection(user string, ca cabby.CollectionAccess) error {
	sql := `update user_collection set can_read = ?, can_write = ? where email = ?`
	args := []interface{}{ca.CanRead, ca.CanWrite, user}

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
            user tu
            inner join user_pass tup
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
						user tu
						inner join user_collection tuc
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

func validateUserPasswordCombo(user cabby.User, password string) error {
	err := user.Validate()
	if err != nil {
		return err
	}
	return validatePassword(password)
}

func validateUserCollection(user string, ca cabby.CollectionAccess) error {
	if user == "" {
		return fmt.Errorf("User undefined")
	}
	if ca.ID.IsEmpty() {
		return fmt.Errorf("Invalid collection ID")
	}
	return nil
}
