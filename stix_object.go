package main

import (
	"encoding/json"
	"errors"

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

type stixObjects struct {
	Objects [][]byte
}

func (s *stixObjects) read(ts taxiiStorer, tf taxiiFilter) (result taxiiResult, err error) {
	sos := *s

	if len(tf.stixID) > 0 {
		result, err = ts.read("stixObject", []interface{}{tf.collectionID, tf.stixID}, tf)
	} else {
		result, err = ts.read("stixObjects", []interface{}{tf.collectionID}, tf)
	}

	if err != nil {
		return
	}

	sos = result.data.(stixObjects)
	*s = sos
	return
}

/* helpers */

func bytesToStixObject(b []byte) (stixObject, error) {
	var so stixObject
	err := json.Unmarshal(b, &so)
	if err != nil {
		return stixObject{}, err
	}

	so.ID, err = s.MarshalStixID(so.RawID)
	so.Object = b
	return so, err
}

func stixObjectsToBundle(sos stixObjects) (s.Bundle, error) {
	b, err := s.NewBundle()
	if err != nil {
		return b, err
	}

	for _, o := range sos.Objects {
		b.Objects = append(b.Objects, o)
	}

	if len(b.Objects) == 0 {
		log.Warn("Can't return an empty bundle, returning error to caller")
		return b, errors.New("No data returned, empty bundle")
	}
	return b, err
}
