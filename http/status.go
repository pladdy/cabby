package http

import (
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
)

// StatusHandler holds a cabby StatusService
type StatusHandler struct {
	StatusService cabby.StatusService
}

// Get serves a status resource
func (h StatusHandler) Get(w http.ResponseWriter, r *http.Request) {
	status, err := h.StatusService.Status(r.Context(), takeStatusID(r))
	if err != nil {
		internalServerError(w, err)
		return
	}

	if status.TotalCount <= 0 {
		resourceNotFound(w, errors.New("No status available for this id"))
		return
	}

	writeContent(w, r, cabby.TaxiiContentType, resourceToJSON(status))
}

// Post handles post request
func (h StatusHandler) Post(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Allow", "Get, Head")
	methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
}
