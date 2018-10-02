package http

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

func noResources(w http.ResponseWriter, resources int, cr cabby.Range) bool {
	err := errors.New("No resources available for this request")

	if cr.Set && resources <= 0 {
		rangeNotSatisfiable(w, err, cr)
		return true
	}
	if resources <= 0 {
		resourceNotFound(w, err)
		return true
	}
	return false
}

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.WithFields(log.Fields{"value": v, "error": err}).Panic("Can't convert to JSON")
	}
	return string(b)
}

func withHSTS(w http.ResponseWriter) http.ResponseWriter {
	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
	return w
}

func writeContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	io.WriteString(w, content)
}

func writePartialContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(http.StatusPartialContent)
	io.WriteString(w, content)
}
