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

	if len(tu.CollectionAccess) <= 0 {
		err = errors.New("No access to any collections")
	}
	return tu, err
}

func (tu *taxiiUser) read(pass string) error {
	user := *tu

	ts, err := newTaxiiStorer()
	if err != nil {
		return err
	}

	query, err := ts.parse("read", "taxiiUser")
	if err != nil {
		return err
	}

	result, err := ts.read(query, "taxiiUser", []interface{}{tu.Email, pass})
	tcas := result.([]taxiiCollectionAccess)

	for _, tca := range tcas {
		user.CollectionAccess[tca.ID] = tca
	}
	*tu = user

	return err
}
