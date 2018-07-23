package main

import (
	"crypto/sha256"
	"fmt"
)

type taxiiCollectionAccess struct {
	ID       taxiiID `json:"id"`
	CanRead  bool    `json:"can_read"`
	CanWrite bool    `json:"can_write"`
}

type taxiiUser struct {
	Email                string                            `json:"email"`
	CanAdmin             bool                              `json:"can_admin"`
	CollectionAccessList map[taxiiID]taxiiCollectionAccess `json:"collection_access_list"`
}

func newTaxiiUser(ts taxiiStorer, u, p string) (taxiiUser, error) {
	tu := taxiiUser{Email: u, CollectionAccessList: make(map[taxiiID]taxiiCollectionAccess)}
	err := tu.read(ts, hash(p))
	return tu, err
}

func (tu *taxiiUser) create(ts taxiiStorer) error {
	return createResource(ts, "taxiiUser", []interface{}{tu.Email, tu.CanAdmin})
}

func (tu *taxiiUser) delete(ts taxiiStorer) error {
	err := ts.delete("taxiiUser", []interface{}{tu.Email})
	if err != nil {
		return err
	}

	return ts.delete("taxiiUserPassword", []interface{}{tu.Email})
}

func (tu *taxiiUser) read(ts taxiiStorer, hashedPass string) error {
	user := *tu

	result, err := ts.read("taxiiUser", []interface{}{tu.Email, hashedPass})
	if err != nil {
		return err
	}

	user = result.data.(taxiiUser)

	err = addCollectionsToUser(ts, tu)
	if err != nil {
		return err
	}

	*tu = user
	return err
}

func (tu *taxiiUser) update(ts taxiiStorer) error {
	return ts.update("taxiiUser", []interface{}{tu.CanAdmin, tu.Email})
}

func (tu *taxiiUser) valid() bool {
	if tu.Email == "" {
		return false
	}
	return true
}

type taxiiUserCollection struct {
	Email                 string `json:"email"`
	taxiiCollectionAccess `json:"collection_access"`
}

func (tuc *taxiiUserCollection) create(ts taxiiStorer) error {
	tca := tuc.taxiiCollectionAccess
	return createResource(ts,
		"taxiiUserCollection",
		[]interface{}{tuc.Email, tca.ID.String(), tca.CanRead, tca.CanWrite})
}

func (tuc *taxiiUserCollection) delete(ts taxiiStorer) error {
	return ts.delete("taxiiUserCollection", []interface{}{tuc.Email, tuc.taxiiCollectionAccess.ID.String()})
}

func (tuc *taxiiUserCollection) read(ts taxiiStorer) error {
	userCollection := *tuc

	result, err := ts.read("taxiiUserCollection", []interface{}{tuc.Email, tuc.taxiiCollectionAccess.ID.String()})
	if err != nil {
		return err
	}

	userCollection = result.data.(taxiiUserCollection)

	*tuc = userCollection
	return err
}

func (tuc *taxiiUserCollection) update(ts taxiiStorer) error {
	tca := tuc.taxiiCollectionAccess
	return ts.update("taxiiUserCollection", []interface{}{tca.CanRead, tca.CanWrite, tuc.Email, tca.ID.String()})
}

type taxiiUserPassword struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func (tup *taxiiUserPassword) create(ts taxiiStorer) error {
	return createResource(ts, "taxiiUserPassword", []interface{}{tup.Email, hash(tup.Password)})
}

func (tup *taxiiUserPassword) update(ts taxiiStorer) error {
	return ts.update("taxiiUserPassword", []interface{}{hash(tup.Password), tup.Email})
}

/* helpers */

func addCollectionsToUser(ts taxiiStorer, tu *taxiiUser) error {
	tcas, err := assignedCollections(ts, tu.Email)
	if err != nil {
		return err
	}

	tu.CollectionAccessList = make(map[taxiiID]taxiiCollectionAccess)
	for _, tca := range tcas {
		tu.CollectionAccessList[tca.ID] = tca
	}
	return err
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

func hash(pass string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(pass)))
}
