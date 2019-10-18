package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// ObjectMethods lists allowed methods
const ObjectMethods = "Get, Delete, Head"

// ObjectHandler handles Objects requests
type ObjectHandler struct {
	ObjectService cabby.ObjectService
}

//Delete handles a delete of an object; can only be done given an ID
func (h ObjectHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectHandler"}).Debug("Handler called")

	err := h.ObjectService.DeleteObject(r.Context(), takeCollectionID(r), takeObjectID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}
	writeContent(w, r, cabby.TaxiiContentType, "")
}

// Get handles a get request for the objects endpoint
func (h ObjectHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectHandler", "objectID": takeObjectID(r)}).Debug("Handler called")

	if !verifySupportedMimeType(w, r, "Accept", cabby.TaxiiContentType) {
		return
	}

	h.getObject(w, r)
}

func (h ObjectHandler) getObject(w http.ResponseWriter, r *http.Request) {
	objects, err := h.ObjectService.Object(r.Context(), takeCollectionID(r), takeObjectID(r), newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(objects) <= 0 {
		resourceNotFound(w, errors.New("No objects defined in this collection"))
		return
	}

	envelope := objectsToEnvelope(objects, cabby.Range{})
	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(envelope))
}

// Post handler
func (h ObjectHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", VersionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
