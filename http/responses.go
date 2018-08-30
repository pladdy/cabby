package http

import (
	"encoding/json"
	"io"
	"net/http"

	log "github.com/sirupsen/logrus"
)

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
