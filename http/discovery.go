package http

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strconv"

	cabby "github.com/pladdy/cabby2"
)

// DiscoveryHandler holds a cabby DiscoveryService
type DiscoveryHandler struct {
	DiscoveryService cabby.DiscoveryService
}

// HandleDiscovery serves a discovery resource
func (h *DiscoveryHandler) HandleDiscovery(port int) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		if !requestMethodIsGet(r) {
			methodNotAllowed(w, fmt.Errorf("Invalid method: %s", r.Method))
			return
		}

		result, err := h.DiscoveryService.Read()
		if err != nil {
			internalServerError(w, err)
			return
		}

		discovery, ok := result.Data.(cabby.Discovery)
		if !ok {
			internalServerError(w, errors.New("Invalid result"))
			return
		}

		discovery.Default = insertPort(discovery.Default, port)

		for i := 0; i < len(discovery.APIRoots); i++ {
			discovery.APIRoots[i] = swapPath(discovery.Default, discovery.APIRoots[i]) + "/"
		}

		if discovery.Title == "" {
			resourceNotFound(w, errors.New("Discovery not defined"))
		} else {
			writeContent(w, TaxiiContentType, resourceToJSON(discovery))
		}
	})
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
