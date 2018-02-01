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

func (tu *taxiiUser) read(pass string) error {
	user := *tu

	ts, err := newTaxiiStorer()
	if err != nil {
		fail.Println(err)
		return err
	}

	valid, err := verifyValidUser(ts, tu.Email, pass)
	if !valid || err != nil {
		fail.Println(err)
		return err
	}

	tcas, err := assignedCollections(ts, tu.Email)
	if err != nil {
		fail.Println(err)
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

	tq, err := ts.parse("read", "taxiiCollectionAccess")
	if err != nil {
		fail.Println(err)
		return tcas, err
	}

	result, err := ts.read(tq, []interface{}{e})
	if err != nil {
		return tcas, err
	}

	tcas = result.([]taxiiCollectionAccess)
	return tcas, err
}

func verifyValidUser(ts taxiiStorer, e, p string) (bool, error) {
	var valid bool

	tq, err := ts.parse("read", "taxiiUser")
	if err != nil {
		fail.Println(err)
		return valid, err
	}

	result, err := ts.read(tq, []interface{}{e, p})
	valid = result.(bool)

	if valid != true {
		fail.Println(err)
		err = errors.New("Invalid user")
	}
	return valid, err
}
