package tester

import (
	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

// GenerateObject generates a STIX object given a type
func GenerateObject(objectType string) cabby.Object {
	obj := Object
	id, _ := stones.NewIdentifier(objectType)
	obj.ID = id
	return obj
}
