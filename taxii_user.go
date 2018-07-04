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
	CanAdmin         bool
	CollectionAccess map[taxiiID]taxiiCollectionAccess
}

func newTaxiiUser(ts taxiiStorer, u, p string) (taxiiUser, error) {
	tu := taxiiUser{Email: u, CollectionAccess: make(map[taxiiID]taxiiCollectionAccess)}
	err := tu.read(ts, fmt.Sprintf("%x", sha256.Sum256([]byte(p))))
	return tu, err
}

func (tu *taxiiUser) create(ts taxiiStorer, pass string) error {
	var err error

	parts := []struct {
		resource string
		args     []interface{}
	}{
		{"taxiiUser", []interface{}{tu.Email, tu.CanAdmin}},
		{"taxiiUserPass", []interface{}{tu.Email, pass}},
	}

	for _, p := range parts {
		err = createResource(ts, p.resource, p.args)
		if err != nil {
			return err
		}
	}

	return err
}

func (tu *taxiiUser) read(ts taxiiStorer, pass string) error {
	user := *tu

	result, err := ts.read("taxiiUser", []interface{}{tu.Email, pass})
	if err != nil {
		return err
	}

	user, ok := result.data.(taxiiUser)
	if !ok {
		return errors.New("Invalid user")
	}

	tcas, err := assignedCollections(ts, tu.Email)
	if err != nil {
		return err
	}

	// add collections to user object
	user.CollectionAccess = make(map[taxiiID]taxiiCollectionAccess)
	for _, tca := range tcas {
		user.CollectionAccess[tca.ID] = tca
	}

	*tu = user
	return err
}

func (tu *taxiiUser) valid() bool {
	if tu.Email == "" {
		return false
	}
	return true
}

func assignedCollections(ts taxiiStorer, u string) ([]taxiiCollectionAccess, error) {
	var tcas []taxiiCollectionAccess

	result, err := ts.read("taxiiCollectionAccess", []interface{}{u})
	if err != nil {
		return tcas, err
	}

	tcas = result.data.([]taxiiCollectionAccess)
	return tcas, err
}
