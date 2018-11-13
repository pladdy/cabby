package tester

import (
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

// GenerateObject generates a STIX object given a type
func GenerateObject(objectType string) stones.Object {
	obj := Object
	id, err := stones.NewIdentifier(objectType)
	if err != nil {
		log.Error("Something went wrong creating the test object")
	}

	obj.ID = id
	return obj
}
