package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// VersionsMethods lists allowed methods
const VersionsMethods = "Get, Head"

// VersionsHandler holds a cabby VersionsService
type VersionsHandler struct {
	VersionsService cabby.VersionsService
}

// Delete handler
func (h VersionsHandler) Delete(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", VersionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}

// Get serves a Versions resource
func (h VersionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "VersionsHandler"}).Debug("Handler called")

	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err, cr)
		return
	}

	versions, err := h.VersionsService.Versions(r.Context(), takeCollectionID(r), takeObjectID(r), &cr, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(w, len(versions.Versions), cr) {
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", cr.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", cr.AddedAfterLast())

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, r, cabby.TaxiiContentType, resourceToJSON(versions))
	} else {
		writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(versions))
	}
}

// Post handler
func (h VersionsHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", VersionsMethods)
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
