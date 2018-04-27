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

type stixObjects struct {
	Objects [][]byte
}

func (s *stixObjects) read(ts taxiiStorer, collectionID string, stixID ...string) error {
	sos := *s

	var result taxiiResult
	var err error

	if len(stixID) > 0 {
		result, err = ts.read("stixObject", []interface{}{collectionID, stixID[0]})
	} else {
		result, err = ts.read("stixObjects", []interface{}{collectionID})
	}

	if err != nil {
		return err
	}
	sos = result.data.(stixObjects)

	*s = sos
	return err
}

/* helper */

func writeBundle(b s.Bundle, cid string, ts taxiiStorer) {
	writeErrs := make(chan error, len(b.Objects))
	writes := make(chan interface{}, minBuffer)

	go ts.create("stixObject", writes, writeErrs)

	for _, object := range b.Objects {
		so, err := newStixObject(object)
		if err != nil {
			writeErrs <- err
			continue
		}
		log.WithFields(log.Fields{"stix_id": so.RawID}).Info("Sending to data store")
		writes <- []interface{}{so.RawID, so.Type, so.Created, so.Modified, so.Object, cid}
	}

	close(writes)

	// is this dumb?  errors are logged in the taxiiStorer...what's the point of having them here?
	// ie: do i need to be passing an error channel around?
	for e := range writeErrs {
		log.Error(e)
	}
}
