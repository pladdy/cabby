package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// Object handlers are attached to the ObjectsHandler struct; requests for an object are routed to /objects/ handler
// and below functions are called provided an object id is on path of the request

// ObjectMethods lists allowed methods
const ObjectMethods = "Get, Delete, Head"

/* Delete */

//Delete handles a delete of an object; can only be done given an ID
func (h ObjectsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ObjectsHandler"}).Debug("Handler called")

	if takeObjectID(r) == "" {
		w.Header().Set("Allow", ObjectsMethods)
		methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
		return
	}

	err := h.ObjectService.DeleteObject(r.Context(), takeCollectionID(r), takeObjectID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}
	writeContent(w, r, cabby.TaxiiContentType, "")
}

/* Get */

func (h ObjectsHandler) getObject(w http.ResponseWriter, r *http.Request) {
	objects, err := h.ObjectService.Object(r.Context(), takeCollectionID(r), takeObjectID(r), newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(objects) <= 0 {
		resourceNotFound(w, errors.New("No objects defined in this collection"))
		return
	}

	envelope, err := objectsToEnvelope(objects)
	if err != nil {
		internalServerError(w, errors.New("Unable to create envelope"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(envelope))
}
