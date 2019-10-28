package http

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/pladdy/cabby"
	"github.com/pladdy/stones"
	log "github.com/sirupsen/logrus"
)

func noResources(resources int) bool {
	if resources <= 0 {
		return true
	}
	return false
}

func objectsToEnvelope(objects []stones.Object, p cabby.Page) (e cabby.Envelope) {
	if int(p.Total) > len(objects) {
		e.More = true
	}

	for _, o := range objects {
		e.Objects = append(e.Objects, json.RawMessage(o.Source))
	}
	return
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

func write(w http.ResponseWriter, r *http.Request, content string) {
	if r.Method == http.MethodHead {
		content = ""
	}

	bytes, err := io.WriteString(w, content)
	if err != nil {
		log.WithFields(log.Fields{"bytes": bytes, "content": content, "error": err}).Error(
			"Error trying to write resource to the response",
		)
	}

	*r = *r.WithContext(cabby.WithBytes(r.Context(), bytes))
}

func writeContent(w http.ResponseWriter, r *http.Request, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	write(w, r, content)
}
