package http

import (
	"net/http"

	log "github.com/sirupsen/logrus"
)

// func registerAPIRoots(ds cabby.DataStore, sm *http.ServeMux) {
// 	apiRoots := cabby.APIRoots{}
// 	err := apiRoots.read(ds)
// 	if err != nil {
// 		log.Error("Unable to register api roots")
// 		return handler, err
// 	}
//
// 	for _, rootPath := range apiRoots.RootPaths {
// 		registerAPIRoot(ds, rootPath, sm)
// 	}
// }
//
// func registerAPIRoot(ds cabby.DataStore, rootPath string, sm *http.ServeMux) {
// 	ar := cabby.APIRoot{Path: rootPath}
// 	err := ar.read(ds)
// 	if err != nil {
// 		log.WithFields(log.Fields{"api_root": rootPath}).Error("Unable to read API roots")
// 		return
// 	}
//
// 	if rootPath != "" {
// 		registerCollectionRoutes(ds, ar, rootPath, sm)
// 		registerRoute(sm, rootPath+"/collections", withAcceptTaxii(handleTaxiiCollections(ds)))
// 		registerRoute(sm, rootPath+"/status", withAcceptTaxii(handleTaxiiStatus(ds)))
// 		registerRoute(sm, rootPath, withAcceptTaxii(handleTaxiiAPIRoot(ds)))
// 	}
// }
//
// func registerCollectionRoutes(ds cabby.DataStore, ar taxiiAPIRoot, rootPath string, sm *http.ServeMux) {
// 	rcs := routableCollections{}
// 	err := rcs.read(ds, rootPath)
// 	if err != nil {
// 		log.WithFields(log.Fields{"api_root": rootPath}).Error("Unable to read routable collections")
// 	}
//
// 	for _, collectionID := range rcs.CollectionIDs {
// 		registerRoute(sm,
// 			rootPath+"/collections/"+collectionID.String()+"/objects",
// 			withAcceptStix(handleTaxiiObjects(ds, ar.MaxContentLength)))
// 		registerRoute(sm,
// 			rootPath+"/collections/"+collectionID.String()+"/manifest",
// 			withAcceptTaxii(handleTaxiiManifest(ds)))
// 	}
// }

// RegisterRoute takes a handler, a path, and a handler functino to associate
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
// 	// registerAPIRoots(ds, handler)
// 	//
// 	// // admin routes
// 	// registerRoute(handler, "admin/api_root", withAcceptTaxii(handleAdminTaxiiAPIRoot(ds, handler)))
// 	// registerRoute(handler, "admin/collections", withAcceptTaxii(handleAdminTaxiiCollections(ds)))
// 	// registerRoute(handler, "admin/discovery", withAcceptTaxii(handleAdminTaxiiDiscovery(ds)))
// 	// registerRoute(handler, "admin/user", withAcceptTaxii(handleAdminTaxiiUser(ds)))
// 	// registerRoute(handler, "admin/user/collection", withAcceptTaxii(handleAdminTaxiiUserCollection(ds)))
// 	// registerRoute(handler, "admin/user/password", withAcceptTaxii(handleAdminTaxiiUserPassword(ds)))
//
// 	registerRoute(handler, "taxii", withAcceptTaxii(dh.HandleDiscovery(port)))
// 	registerRoute(handler, "/", handleUndefinedRequest)
// 	return handler, err
// }
