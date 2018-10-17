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
		rangeNotSatisfiable(w, err, cr)
		return
	}

	collections, err := h.CollectionService.Collections(r.Context(), takeAPIRoot(r), &cr)
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(w, len(collections.Collections), cr) {
		return
	}

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, r, cabby.TaxiiContentType, resourceToJSON(collections))
	} else {
		writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(collections))
	}
}

// Post handles post request
func (h CollectionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "Get, Head")
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
