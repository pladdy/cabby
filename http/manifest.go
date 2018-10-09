package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// ManifestHandler holds a cabby ManifestService
type ManifestHandler struct {
	ManifestService cabby.ManifestService
}

// Get serves a manifest resource
func (h ManifestHandler) Get(w http.ResponseWriter, r *http.Request) {
	log.WithFields(log.Fields{"handler": "ManifestHandler"}).Debug("Handler called")

	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err, cr)
		return
	}

	manifest, err := h.ManifestService.Manifest(r.Context(), takeCollectionID(r), &cr, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if noResources(w, len(manifest.Objects), cr) {
		return
	}

	w.Header().Set("X-TAXII-Date-Added-First", cr.AddedAfterFirst())
	w.Header().Set("X-TAXII-Date-Added-Last", cr.AddedAfterLast())

	if cr.Set {
		w.Header().Set("Content-Range", cr.String())
		writePartialContent(w, cabby.TaxiiContentType, resourceToJSON(manifest))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(manifest))
	}
}

// Post handles post request
func (h ManifestHandler) Post(w http.ResponseWriter, r *http.Request) {
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
