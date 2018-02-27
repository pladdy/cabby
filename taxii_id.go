package main

import uuid "github.com/satori/go.uuid"

type taxiiID struct {
	uuid.UUID
}

func newTaxiiID(arg ...string) (taxiiID, error) {
	if len(arg) > 0 && len(arg[0]) > 0 {
		id, err := uuid.FromString(arg[0])
		return taxiiID{id}, err
	}

	id, err := uuid.NewV4()
	return taxiiID{id}, err
}

func (ti *taxiiID) isEmpty() bool {
	empty := &taxiiID{}
	if ti.String() == empty.String() {
		return true
	}
	return false
}
