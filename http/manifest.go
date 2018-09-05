package http

import (
	"errors"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// ManifestHandler holds a cabby ManifestService
type ManifestHandler struct {
	ManifestService cabby.ManifestService
}

// Get serves a manifest resource
func (h ManifestHandler) Get(w http.ResponseWriter, r *http.Request) {
	cr, err := cabby.NewRange(r.Header.Get("Range"))
	if err != nil {
		rangeNotSatisfiable(w, err)
		return
	}

	manifest, err := h.ManifestService.Manifest(takeCollectionID(r), &cr, newFilter(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(manifest.Objects) <= 0 {
		resourceNotFound(w, errors.New("No manifest available for this collection"))
		return
	}

	if cr.Valid() {
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
