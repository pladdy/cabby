package main

import (
	"encoding/json"

	s "github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

type stixObject struct {
	RawID        string   `json:"id"`
	ID           s.StixID `json:"-"`
	Type         string   `json:"type"`
	Created      string   `json:"created"`
	Modified     string   `json:"modified"`
	Object       []byte
	CollectionID taxiiID
}

func newStixObject(b []byte) (stixObject, error) {
	var so stixObject
	err := json.Unmarshal(b, &so)
	if err != nil {
		return stixObject{}, err
	}

	so.ID, err = s.MarshalStixID(so.RawID)
	so.Object = b
	return so, err
}

/* helper */

func writeBundle(b s.Bundle, cid string, ts taxiiStorer) <-chan error {
	readErrs := make(chan error, len(b.Objects))
	writes := make(chan interface{}, minBuffer)

	go ts.create("stixObject", writes, readErrs)

	for _, object := range b.Objects {
		so, err := newStixObject(object)
		if err != nil {
			readErrs <- err
		}
		log.Info("Writing:", so.RawID)
		writes <- []interface{}{so.RawID, so.Type, so.Created, so.Modified, so.Object, cid}
	}

	close(writes)

	return readErrs
}
