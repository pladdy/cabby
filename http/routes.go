package http

import (
	"net/http"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
)

func registerAPIRoots(ds cabby.DataStore, sm *http.ServeMux) {
	ah := APIRootHandler{APIRootService: ds.APIRootService()}
	apiRoots, err := ah.APIRootService.APIRoots()
	if err != nil {
		log.Error("Unable to register api roots")
		return
	}

	for _, apiRoot := range apiRoots {
		registerAPIRoot(ah, apiRoot.Path, sm)
		registerCollectionRoutes(ds, apiRoot.Path, sm)
	}
}

func registerAPIRoot(ah APIRootHandler, path string, sm *http.ServeMux) {
	if path != "" {
		registerRoute(sm, path, WithAcceptType(RouteRequest(ah), cabby.TaxiiContentType))
	}
}

func registerCollectionRoutes(ds cabby.DataStore, path string, sm *http.ServeMux) {
	ch := CollectionsHandler{CollectionService: ds.CollectionService()}

	registerRoute(sm, path+"/collections", WithAcceptType(RouteRequest(ch), cabby.TaxiiContentType))

	acs, err := ch.CollectionService.CollectionsInAPIRoot(path)
	if err != nil {
		log.WithFields(log.Fields{"api_root": path}).Error("Unable to read collections")
	}

	for _, id := range acs.CollectionIDs {
		registerRoute(
			sm,
			path+"/collections/"+id.String(),
			WithAcceptType(RouteRequest(ch), cabby.TaxiiContentType))
		// registerRoute(sm,
		// 	path+"/collections/"+collection.ID.String()+"/objects",
		// 	withAcceptStix(handleTaxiiObjects(ds, ar.MaxContentLength)))
		// registerRoute(sm,
		// 	path+"/collections/"+collection.ID.String()+"/manifest",
		// 	withAcceptTaxii(handleTaxiiManifest(ds)))
	}
	//registerRoute(sm, ar.Path+"/status", withAcceptTaxii(handleTaxiiStatus(ds)))
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
