package main

import (
	"crypto/tls"
	"flag"
	"net/http"
	"os"
	"strconv"

	log "github.com/sirupsen/logrus"
)

const (
	cabbyEnvironmentVariable = "CABBY_ENVIRONMENT"
	defaultCabbyEnvironment  = "development"
	defaultDevelopmentConfig = "config/cabby.json"
	defaultProductionConfig  = "/etc/cabby/cabby.json"
)

var cabbyConfigs = map[string]string{
	"development": defaultDevelopmentConfig,
	"production":  defaultProductionConfig,
}

func newCabby(c config) (*http.Server, error) {
	var server http.Server

	ts, err := newTaxiiStorer(c.DataStore["name"], c.DataStore["path"])
	if err != nil {
		return &server, err
	}

	handler, err := setupHandler(ts, c.Port)
	if err != nil {
		return &server, err
	}
	return setupServer(ts, handler, c), err
}

func registerAPIRoot(ts taxiiStorer, rootPath string, sm *http.ServeMux) {
	ar := taxiiAPIRoot{}
	err := ar.read(ts, rootPath)
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

	sm.HandleFunc(route, withRequestLogging(h))
}

func setupHandler(ts taxiiStorer, port int) (*http.ServeMux, error) {
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

	registerRoute(handler, "taxii", withAcceptTaxii(handleTaxiiDiscovery(ts, port)))
	registerRoute(handler, "/", handleUndefinedRequest)
	return handler, err
}

// server is set up with basic auth and HSTS applied to each handler
func setupServer(ts taxiiStorer, h http.Handler, c config) *http.Server {
	p := strconv.Itoa(c.Port)
	log.WithFields(log.Fields{"port": p}).Info("Server port configured")

	return &http.Server{
		Addr:         ":" + p,
		Handler:      withBasicAuth(h, ts),
		TLSConfig:    setupTLS(),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
}

func setupTLS() *tls.Config {
	return &tls.Config{
		MinVersion:               tls.VersionTLS12,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}
}

func main() {
	cabbyEnv := os.Getenv(cabbyEnvironmentVariable)
	if len(cabbyEnv) == 0 {
		cabbyEnv = defaultCabbyEnvironment
	}

	// set up flag, but don't use; overwrite with defalut dev config path for now
	var configPath = flag.String("config", cabbyConfigs[cabbyEnv], "path to cabby config file")
	flag.Parse()

	cs := configs{}.parse(*configPath)
	c := cs[cabbyEnv]

	server, err := newCabby(c)
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("Can't start server")
	}

	log.Fatal(server.ListenAndServeTLS(c.SSLCert, c.SSLKey))
}
