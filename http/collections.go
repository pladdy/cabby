package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// CollectionsMethods lists allowed methods
const CollectionsMethods = "Get, Head"

// CollectionsHandler handles Collections requests
type CollectionsHandler struct {
	CollectionService cabby.CollectionService
}

// Delete handler
func (h CollectionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", CollectionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}

// Get handles a get request
func (h CollectionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "CollectionsHandler"}).Debug("Handler called")

	if len(takeCollectionID(r)) > 0 {
		resourceNotFound(w, errors.New("Collection ID doesn't exist in this API Root"))
		return
	}

	p, err := cabby.NewPage(takeLimit(r))
	if err != nil {
		badRequest(w, err)
		return
	}

	collections, err := h.CollectionService.Collections(r.Context(), takeAPIRoot(r), &p)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(len(collections.Collections)) {
		resourceNotFound(w, errors.New("No resources available for this request"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(collections))
}

// Post handler
func (h CollectionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", CollectionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
