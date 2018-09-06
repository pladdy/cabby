package http

import (
	"errors"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// CollectionHandler handles Collection requestion
type CollectionHandler struct {
	CollectionService cabby.CollectionService
}

// Get handles a get request
func (h CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	collection, err := h.CollectionService.Collection(r.Context(), takeUser(r), takeAPIRoot(r), takeCollectionID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if collection.ID.IsEmpty() {
		resourceNotFound(w, errors.New("Collection ID doesn't exist in this API Root"))
		return
	}

	writeContent(w, cabby.TaxiiContentType, resourceToJSON(collection))
}

// Post handles post request
func (h CollectionHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
