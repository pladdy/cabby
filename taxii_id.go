package main

import uuid "github.com/satori/go.uuid"

type taxiiID struct {
	uuid.UUID
}

const cabbyTaxiiNamespace = "15e011d3-bcec-4f41-92d0-c6fc22ab9e45"

func taxiiIDFromString(s string) (taxiiID, error) {
	id, err := uuid.FromString(s)
	return taxiiID{id}, err
}

func taxiiIDUsingString(s string) (taxiiID, error) {
	ns, err := uuid.FromString(cabbyTaxiiNamespace)
	if err != nil {
		return taxiiID{}, err
	}

	id := uuid.NewV5(ns, s)
	return taxiiID{id}, err
}

// creates a V4 UUID and returns it as a taxiiID
func newTaxiiID() (taxiiID, error) {
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
