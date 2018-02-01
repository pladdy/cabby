package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const (
	minBuffer          = 10
	configPath         = "config/cabby.json"
	stixContentType20  = "application/vnd.oasis.stix+json; version=2.0"
	stixContentType    = "application/vnd.oasis.stix+json"
	taxiiContentType20 = "application/vnd.oasis.taxii+json; version=2.0"
	taxiiContentType   = "application/vnd.oasis.taxii+json"
)

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func init() {
	config = cabbyConfig{}.parse(configPath)
}

func newCabby() *http.Server {
	handler := setupHandler()
	return setupServer(handler)
}

func registerAPIRoot(apiRoot string, h *http.ServeMux) {
	u, err := url.Parse(apiRoot)
	if u.Path == "" || err != nil {
		warn.Panic(err)
	}

	info.Println("Registering handler for", u.String()+"collections/")
	h.HandleFunc(u.Path+"collections/", handleTaxiiCollections)

	info.Println("Registering handler for", u.String())
	h.HandleFunc(u.Path, handleTaxiiAPIRoot)
}

func setupHandler() *http.ServeMux {
	handler := http.NewServeMux()

	for _, apiRoot := range config.Discovery.APIRoots {
		if config.validAPIRoot(apiRoot) {
			registerAPIRoot(apiRoot, handler)
		}
	}

	handler.HandleFunc("/taxii/", handleTaxiiDiscovery)
	handler.HandleFunc("/", handleUndefinedRequest)

	return handler
}

// server is set up with basicAuth and HSTS applied to each handler
func setupServer(h http.Handler) *http.Server {
	port := strconv.Itoa(config.Port)
	info.Println("Server will listen on port " + port)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      basicAuth(h),
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
	server := newCabby()
	fail.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
