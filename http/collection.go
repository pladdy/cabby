package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// CollectionHandler handles Collection requestion
type CollectionHandler struct {
	CollectionService cabby.CollectionService
}

// Get handles a get request
func (h CollectionHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "CollectionHandler"}).Debug("Handler called")

	collection, err := h.CollectionService.Collection(r.Context(), takeAPIRoot(r), takeCollectionID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if collection.ID.IsEmpty() {
		resourceNotFound(w, errors.New("Collection ID doesn't exist in this API Root"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(collection))
}

// Post handles post request
func (h CollectionHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "Get, Head")
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
