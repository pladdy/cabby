package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"
)

const defaultConfig = "config/cabby.json"

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func newCabby(configPath string) (*http.Server, error) {
	var server http.Server

	config = cabbyConfig{}.parse(configPath)

	ts, err := newTaxiiStorer(config.DataStore["name"], config.DataStore["path"])
	if err != nil {
		return &server, err
	}

	handler, err := setupHandler(ts)
	if err != nil {
		return &server, err
	}
	return setupServer(ts, handler), err
}

func registerAPIRoot(ts taxiiStorer, rootPath string, h *http.ServeMux) {
	if rootPath != "" {
		path := "/" + rootPath + "/"
		registerRoute(h, path+"collections/", handleTaxiiCollections(ts))
		registerRoute(h, path, handleTaxiiAPIRoot(ts))
	}
}

func registerRoute(sm *http.ServeMux, path string, h http.HandlerFunc) {
	info.Println("Registering handler for", path)
	sm.HandleFunc(path,
		withAcceptTaxii(h))
}

func setupHandler(ts taxiiStorer) (*http.ServeMux, error) {
	handler := http.NewServeMux()

	apiRoots := taxiiAPIRoots{}
	err := apiRoots.read(ts)
	if err != nil {
		fail.Println("Unable to register api roots")
		return handler, err
	}

	for _, rootPath := range apiRoots.RootPaths {
		registerAPIRoot(ts, rootPath, handler)
	}

	registerRoute(handler, "/taxii/", handleTaxiiDiscovery(ts))
	registerRoute(handler, "/", handleUndefinedRequest)
	return handler, err
}

// server is set up with basic auth and HSTS applied to each handler
func setupServer(ts taxiiStorer, h http.Handler) *http.Server {
	port := strconv.Itoa(config.Port)
	info.Println("Server will listen on port " + port)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      withBasicAuth(ts, h),
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
	server, err := newCabby(defaultConfig)
	if err != nil {
		fail.Fatal("Can't start server:", err)
	}

	fail.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
