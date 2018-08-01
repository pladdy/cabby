package http

import (
	"encoding/json"
	"io"
	"net/http"
	"strings"

	log "github.com/sirupsen/logrus"
)

// func withHSTS(w http.ResponseWriter) http.ResponseWriter {
// 	w.Header().Add("Strict-Transport-Security", "max-age="+sixMonthsOfSeconds+"; includeSubDomains")
// 	return w
// }

func getToken(s string, i int) string {
	tokens := strings.Split(s, "/")

	if len(tokens) > i {
		return tokens[i]
	}
	return ""
}

func getAPIRoot(p string) string {
	var rootIndex = 1
	return getToken(p, rootIndex)
}

func lastURLPathToken(u string) string {
	u = strings.TrimSuffix(u, "/")
	tokens := strings.Split(u, "/")
	return tokens[len(tokens)-1]
}

func resourceToJSON(v interface{}) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.WithFields(log.Fields{
			"value": v,
			"error": err,
		}).Panic("Can't convert to JSON")
	}
	return string(b)
}

func writeContent(w http.ResponseWriter, contentType, content string) {
	w.Header().Set("Content-Type", contentType)
	io.WriteString(w, content)
}

// func writePartialContent(w http.ResponseWriter, contentType, content string) {
// 	w.Header().Set("Content-Type", contentType)
// 	w.WriteHeader(http.StatusPartialContent)
// 	io.WriteString(w, content)
// }
