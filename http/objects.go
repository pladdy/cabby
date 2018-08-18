package http

import (
	"errors"
	"fmt"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// ObjectsHandler handles Objects requests
type ObjectsHandler struct {
	ObjectService    cabby.ObjectService
	MaxContentLength int64
}

// Get handles a get request
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
	objectID := takeObjectID(r)

	if objectID == "" {
		h.getObjects(w, r)
		return
	}
	h.getObject(w, r)
}

func (h ObjectsHandler) getObjects(w http.ResponseWriter, r *http.Request) {
	objects, err := h.ObjectService.Objects(takeCollectionID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(objects) <= 0 {
		resourceNotFound(w, errors.New("No objects defined in this collection"))
		return
	}

	writeContent(w, cabby.TaxiiContentType, resourceToJSON(objects))
}

func (h ObjectsHandler) getObject(w http.ResponseWriter, r *http.Request) {
	object, err := h.ObjectService.Object(takeCollectionID(r), takeObjectID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if object.ID == "" {
		resourceNotFound(w, fmt.Errorf("Object ID doesn't exist in this collection"))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(object))
	}
}
