package main

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

func registerAPIRoot(ts taxiiStorer, rootPath string, sm *http.ServeMux) {
	ar := taxiiAPIRoot{Path: rootPath}
	err := ar.read(ts)
	if err != nil {
		log.WithFields(log.Fields{"api_root": rootPath}).Error("Unable to read API roots")
		return
	}

	if rootPath != "" {
		registerCollectionRoutes(ts, ar, rootPath, sm)
		registerRoute(sm, rootPath+"/collections", withAcceptTaxii(handleTaxiiCollections(ts)))
		registerRoute(sm, rootPath+"/status", withAcceptTaxii(handleTaxiiStatus(ts)))
		registerRoute(sm, rootPath, withAcceptTaxii(handleTaxiiAPIRoot(ts)))
	}
}

func registerCollectionRoutes(ts taxiiStorer, ar taxiiAPIRoot, rootPath string, sm *http.ServeMux) {
	rcs := routableCollections{}
	err := rcs.read(ts, rootPath)
	if err != nil {
		log.WithFields(log.Fields{"api_root": rootPath}).Error("Unable to read routable collections")
	}

	for _, collectionID := range rcs.CollectionIDs {
		registerRoute(sm,
			rootPath+"/collections/"+collectionID.String()+"/objects",
			withAcceptStix(handleTaxiiObjects(ts, ar.MaxContentLength)))
		registerRoute(sm,
			rootPath+"/collections/"+collectionID.String()+"/manifest",
			withAcceptTaxii(handleTaxiiManifest(ts)))
	}
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

func setupRouteHandler(ts taxiiStorer, port int) (*http.ServeMux, error) {
	handler := http.NewServeMux()

	apiRoots := taxiiAPIRoots{}
	err := apiRoots.read(ts)
	if err != nil {
		log.Error("Unable to register api roots")
		return handler, err
	}

	for _, rootPath := range apiRoots.RootPaths {
		registerAPIRoot(ts, rootPath, handler)
	}

	// admin routes
	registerRoute(handler, "admin/api_root", withAcceptTaxii(handleAdminTaxiiAPIRoot(ts, handler)))
	registerRoute(handler, "admin/collections", withAcceptTaxii(handleAdminTaxiiCollections(ts)))
	registerRoute(handler, "admin/discovery", withAcceptTaxii(handleAdminTaxiiDiscovery(ts)))
	registerRoute(handler, "admin/user", withAcceptTaxii(handleAdminTaxiiUser(ts)))
	registerRoute(handler, "admin/user/collection", withAcceptTaxii(handleAdminTaxiiUserCollection(ts)))
	registerRoute(handler, "admin/user/password", withAcceptTaxii(handleAdminTaxiiUserPassword(ts)))

	registerRoute(handler, "taxii", withAcceptTaxii(handleTaxiiDiscovery(ts, port)))
	registerRoute(handler, "/", handleUndefinedRequest)
	return handler, err
}
