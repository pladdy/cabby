package http

import (
	"errors"
	"net/http"

	cabby "github.com/pladdy/cabby2"
)

// CollectionsHandler handles Collections requests
type CollectionsHandler struct {
	CollectionService cabby.CollectionService
}

// Get handles a get request
func (h CollectionsHandler) Get(w http.ResponseWriter, r *http.Request) {
	collections, err := h.CollectionService.Collections(takeUser(r), takeAPIRoot(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if len(collections.Collections) <= 0 {
		resourceNotFound(w, errors.New("No collections defined in this API Root"))
		return
	}

	writeContent(w, cabby.TaxiiContentType, resourceToJSON(collections))
}
