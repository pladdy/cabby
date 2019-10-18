package http

import (
	"context"
	"errors"
	"net/http"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

const versionsPathToken = "versions"

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
		registerRoute(sm, path, WithMimeType(routeHandler(ah), "Accept", cabby.TaxiiContentType))
	}
}

func registerCollectionRoutes(ds cabby.DataStore, apiRoot cabby.APIRoot, sm *http.ServeMux) {
	csh := CollectionsHandler{CollectionService: ds.CollectionService()}
	registerRoute(sm, apiRoot.Path+"/collections", WithMimeType(routeHandler(csh), "Accept", cabby.TaxiiContentType))

	ss := ds.StatusService()
	osh := ObjectsHandler{
		MaxContentLength: apiRoot.MaxContentLength,
		ObjectService:    ds.ObjectService(),
		StatusService:    ss}
	mh := ManifestHandler{ManifestService: ds.ManifestService()}
	ch := CollectionHandler{CollectionService: ds.CollectionService()}
	oh := ObjectHandler{ObjectService: ds.ObjectService()}
	vsh := VersionsHandler{VersionsService: ds.VersionsService()}

	acs, err := csh.CollectionService.CollectionsInAPIRoot(context.Background(), apiRoot.Path)
	if err != nil {
		log.WithFields(log.Fields{"api_root": apiRoot.Path}).Error("Unable to read collections")
		return
	}

	for _, collectionID := range acs.CollectionIDs {
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String(),
			WithMimeType(routeHandler(ch), "Accept", cabby.TaxiiContentType))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/objects",
			routeObjectsHandler(oh, osh, vsh))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/manifest",
			WithMimeType(routeHandler(mh), "Accept", cabby.TaxiiContentType))
	}

	sh := StatusHandler{StatusService: ss}
	registerRoute(sm, apiRoot.Path+"/status", WithMimeType(routeHandler(sh), "Accept", cabby.TaxiiContentType))
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

func routeObjectsHandler(oh, osh, vsh RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		log.WithFields(log.Fields{"handler": "routeObjectsHandler"}).Debug("Handler called")
		if takeVersions(r) == versionsPathToken {
			runHandler(vsh, w, r)
			return
		}
		if takeObjectID(r) == "" {
			runHandler(osh, w, r)
			return
		}
		runHandler(oh, w, r)
	}
}

func routeHandler(h RequestHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runHandler(h, w, r)
	}
}

func runHandler(h RequestHandler, w http.ResponseWriter, r *http.Request) {
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
}
