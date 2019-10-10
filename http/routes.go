package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

func registerAPIRoots(ds cabby.DataStore, sm *http.ServeMux) {
	ah := APIRootHandler{APIRootService: ds.APIRootService()}
	apiRoots, err := ah.APIRootService.APIRoots(context.Background())
	if err != nil {
		log.Error("Unable to register api roots")
		return
	}

	for _, apiRoot := range apiRoots {
		registerAPIRoot(ah, apiRoot.Path, sm)
		registerCollectionRoutes(ds, apiRoot, sm)
	}
}

func registerAPIRoot(ah APIRootHandler, path string, sm *http.ServeMux) {
	if path != "" {
		registerRoute(sm, path, WithMimeType(routeRequest(ah), "Accept", cabby.TaxiiContentType))
	}
}

func registerCollectionRoutes(ds cabby.DataStore, apiRoot cabby.APIRoot, sm *http.ServeMux) {
	csh := CollectionsHandler{CollectionService: ds.CollectionService()}
	registerRoute(sm, apiRoot.Path+"/collections", WithMimeType(routeRequest(csh), "Accept", cabby.TaxiiContentType))

	ss := ds.StatusService()
	oh := ObjectsHandler{
		MaxContentLength: apiRoot.MaxContentLength,
		ObjectService:    ds.ObjectService(),
		StatusService:    ss}

	mh := ManifestHandler{ManifestService: ds.ManifestService()}
	ch := CollectionHandler{CollectionService: ds.CollectionService()}

	acs, err := csh.CollectionService.CollectionsInAPIRoot(context.Background(), apiRoot.Path)
	if err != nil {
		log.WithFields(log.Fields{"api_root": apiRoot.Path}).Error("Unable to read collections")
		return
	}

	for _, collectionID := range acs.CollectionIDs {
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String(),
			WithMimeType(routeRequest(ch), "Accept", cabby.TaxiiContentType))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/objects",
			routeRequest(oh))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/manifest",
			WithMimeType(routeRequest(mh), "Accept", cabby.TaxiiContentType))
	}

	sh := StatusHandler{StatusService: ss}
	registerRoute(sm, apiRoot.Path+"/status", WithMimeType(routeRequest(sh), "Accept", cabby.TaxiiContentType))
}

func registerRoute(sm *http.ServeMux, path string, h http.HandlerFunc) {
	// assume route is root
	route := "/"
	if path != "/" {
		route = "/" + path + "/"
	}
	log.WithFields(log.Fields{"route": route}).Info("Registering handler to route")
	sm.HandleFunc(route, h)
}

func routeRequest(h RequestHandler) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer recoverFromPanic(w)

		switch r.Method {
		case http.MethodDelete:
			h.Delete(w, r)
		case http.MethodGet:
			h.Get(w, r)
		case http.MethodPost:
			h.Post(w, r)
		case http.MethodHead:
			// for HEAD requests send to GET, it will omit response
			h.Get(w, r)
		default:
			w.Header().Set("Allow", "Get, Head, Post")
			methodNotAllowed(w, errors.New("HTTP Method "+r.Method+" unrecognized"))
			return
		}
	})
}
