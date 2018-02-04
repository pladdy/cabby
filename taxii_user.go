package main

import (
	"crypto/sha256"
	"errors"
	"fmt"
)

type taxiiCollectionAccess struct {
	ID       taxiiID `json:"id"`
	CanRead  bool    `json:"can_read"`
	CanWrite bool    `json:"can_write"`
}

type taxiiUser struct {
	Email            string
	CollectionAccess map[taxiiID]taxiiCollectionAccess
}

func newTaxiiUser(u, p string) (taxiiUser, error) {
	tu := taxiiUser{Email: u, CollectionAccess: make(map[taxiiID]taxiiCollectionAccess)}
	err := tu.read(fmt.Sprintf("%x", sha256.Sum256([]byte(p))))
	return tu, err
}

func (tu *taxiiUser) create(p string) error {
	var err error

	parts := []struct {
		resource string
		args     []interface{}
	}{
		{"taxiiUser", []interface{}{tu.Email}},
		{"taxiiUserPass", []interface{}{tu.Email, p}},
	}

	for _, p := range parts {
		err = createResource(p.resource, p.args)
		if err != nil {
			return err
		}
	}

	return err
}

func (tu *taxiiUser) read(pass string) error {
	user := *tu

	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	valid, err := verifyValidUser(ts, tu.Email, pass)
	if !valid || err != nil {
		return err
	}

	tcas, err := assignedCollections(ts, tu.Email)
	if err != nil {
		return err
	}

	// add collections to user object
	for _, tca := range tcas {
		user.CollectionAccess[tca.ID] = tca
	}

	*tu = user
	return err
}

func assignedCollections(ts taxiiStorer, e string) ([]taxiiCollectionAccess, error) {
	var tcas []taxiiCollectionAccess

	result, err := readResource("taxiiCollectionAccess", []interface{}{e})
	if err != nil {
		return tcas, err
	}

	tcas = result.([]taxiiCollectionAccess)
	return tcas, err
}

func verifyValidUser(ts taxiiStorer, e, p string) (bool, error) {
	var valid bool

	result, err := readResource("taxiiUser", []interface{}{e, p})
	if err != nil {
		return false, err
	}

	valid = result.(bool)

	if valid != true {
		err = errors.New("Invalid user")
	}
	return valid, err
}
