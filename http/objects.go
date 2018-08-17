package http

import (
	"errors"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// ObjectsHandler handles Objects requests
type ObjectsHandler struct {
	ObjectService cabby.ObjectService
}

// Get handles a get request
func (h ObjectsHandler) Get(w http.ResponseWriter, r *http.Request) {
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
