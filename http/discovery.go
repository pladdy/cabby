package http

import (
	"errors"
	"log"
	"net/http"
	"net/url"
	"strconv"

	cabby "github.com/pladdy/cabby2"
)

// DiscoveryHandler holds a cabby DiscoveryService
type DiscoveryHandler struct {
	DiscoveryService cabby.DiscoveryService
	Port             int
}

// Get serves a discovery resource
func (h DiscoveryHandler) Get(w http.ResponseWriter, r *http.Request) {
	discovery, err := h.DiscoveryService.Discovery()
	if err != nil {
		internalServerError(w, err)
		return
	}

	discovery.Default = insertPort(discovery.Default, h.Port)

	for i := 0; i < len(discovery.APIRoots); i++ {
		discovery.APIRoots[i] = swapPath(discovery.Default, discovery.APIRoots[i]) + "/"
	}

	if discovery.Title == "" {
		resourceNotFound(w, errors.New("Discovery not defined"))
	} else {
		writeContent(w, cabby.TaxiiContentType, resourceToJSON(discovery))
	}
}

/* helpers */

func insertPort(u string, port int) string {
	tokens := urlTokens(u)
	return tokens["scheme"] + "://" + tokens["host"] + ":" + strconv.Itoa(port) + tokens["path"]
}

func swapPath(u, p string) string {
	tokens := urlTokens(u)
	return tokens["scheme"] + "://" + tokens["host"] + "/" + p
}

func urlTokens(u string) map[string]string {
	tokens, err := url.Parse(u)
	if err != nil {
		log.Panic("Can't parse url")
	}
	return map[string]string{"scheme": tokens.Scheme, "host": tokens.Host, "path": tokens.Path}
}
