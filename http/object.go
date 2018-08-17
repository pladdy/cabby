package http

import (
	"fmt"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// ObjectHandler handles Object requestion
type ObjectHandler struct {
	ObjectService cabby.ObjectService
}

// Get handles a get request
func (h ObjectHandler) Get(w http.ResponseWriter, r *http.Request) {
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
