package tester

import (
	"github.com/pladdy/stones"
)

// GenerateObject generates a STIX object given a type
func GenerateObject(objectType string) stones.Object {
	obj := Object
	id, _ := stones.NewIdentifier(objectType)
	obj.ID = id
	return obj
}
