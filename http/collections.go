package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// CollectionsHandler handles Collections requests
type CollectionsHandler struct {
	CollectionService cabby.CollectionService
}

// Get handles a get request
func (h CollectionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "CollectionsHandler"}).Debug("Handler called")

	if len(takeCollectionID(r)) > 0 {
		resourceNotFound(w, errors.New("Collection ID doesn't exist in this API Root"))
		return
	}

	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err)
		return
	}

	collections, err := h.CollectionService.Collections(r.Context(), takeAPIRoot(r), &cr)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(collections.Collections) <= 0 {
		resourceNotFound(w, errors.New("No collections defined in this API Root"))
		return
	}

	if cr.Valid() {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, cabby.TaxiiContentType, resourceToJSON(collections))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(collections))
	}
}

// Post handles post request
func (h CollectionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
