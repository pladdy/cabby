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
	manifest, err := h.ManifestService.Manifest(takeCollectionID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(manifest.Objects) <= 0 {
		resourceNotFound(w, errors.New("No manifest available for this collection"))
		return
	}

	writeContent(w, cabby.TaxiiContentType, resourceToJSON(manifest))
}
