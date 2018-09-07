package http

import (
	"context"
	"net/http"

	cabby "github.com/pladdy/cabby2"
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
		registerRoute(sm, path, WithAcceptType(RouteRequest(ah), cabby.TaxiiContentType))
	}
}

func registerCollectionRoutes(ds cabby.DataStore, apiRoot cabby.APIRoot, sm *http.ServeMux) {
	ch := CollectionsHandler{CollectionService: ds.CollectionService()}
	registerRoute(sm, apiRoot.Path+"/collections", WithAcceptType(RouteRequest(ch), cabby.TaxiiContentType))

	acs, err := ch.CollectionService.CollectionsInAPIRoot(context.Background(), apiRoot.Path)
	if err != nil {
		log.WithFields(log.Fields{"api_root": apiRoot.Path}).Error("Unable to read collections")
		return
	}

	ss := ds.StatusService()
	oh := ObjectsHandler{
		MaxContentLength: apiRoot.MaxContentLength,
		ObjectService:    ds.ObjectService(),
		StatusService:    ss}

	mh := ManifestHandler{ManifestService: ds.ManifestService()}

	for _, collectionID := range acs.CollectionIDs {
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String(),
			WithAcceptType(RouteRequest(ch), cabby.TaxiiContentType))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/objects",
			WithAcceptType(RouteRequest(oh), cabby.StixContentType))
		registerRoute(
			sm,
			apiRoot.Path+"/collections/"+collectionID.String()+"/manifest",
			WithAcceptType(RouteRequest(mh), cabby.TaxiiContentType))
	}

	sh := StatusHandler{StatusService: ss}
	registerRoute(sm, apiRoot.Path+"/status", WithAcceptType(RouteRequest(sh), cabby.TaxiiContentType))
}

func registerRoute(sm *http.ServeMux, path string, h http.HandlerFunc) {
	log.WithFields(log.Fields{"path": path}).Info("Registering handler")

	// assume route is root
	route := "/"
	if path != "/" {
		route = "/" + path + "/"
	}

	sm.HandleFunc(route, h)
}

// // SetupRouteHandler sets up a handler and returns it with routes defined for the server
// func SetupRouteHandler(ds cabby.DataStore, port int) (*http.ServeMux, error) {
// 	handler := http.NewServeMux()
//
// registerAPIRoots(ds, handler)
//
// // admin routes
// registerRoute(handler, "admin/api_root", withAcceptTaxii(handleAdminTaxiiAPIRoot(ds, handler)))
// registerRoute(handler, "admin/collections", withAcceptTaxii(handleAdminTaxiiCollections(ds)))
// registerRoute(handler, "admin/discovery", withAcceptTaxii(handleAdminTaxiiDiscovery(ds)))
// registerRoute(handler, "admin/user", withAcceptTaxii(handleAdminTaxiiUser(ds)))
// registerRoute(handler, "admin/user/collection", withAcceptTaxii(handleAdminTaxiiUserCollection(ds)))
// registerRoute(handler, "admin/user/password", withAcceptTaxii(handleAdminTaxiiUserPassword(ds)))
//
// 	registerRoute(handler, "taxii", withAcceptTaxii(dh.HandleDiscovery(port)))
// 	registerRoute(handler, "/", handleUndefinedRequest)
// 	return handler, err
// }
