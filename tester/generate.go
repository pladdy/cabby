package tester

import (
	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
)

func GenerateObject(objectType string) cabby.Object {
	obj := Object
	id, _ := stones.NewStixID(objectType)
	obj.ID = stones.ID(id.String())
	return obj
}
