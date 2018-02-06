package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
)

const configPath = "config/cabby.json"

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	fail = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func init() {
	config = cabbyConfig{}.parse(configPath)
}

func newCabby() (*http.Server, error) {
	handler, err := setupHandler()
	if err != nil {
		fail.Println(err)
	}
	return setupServer(handler), err
}

func registerAPIRoot(apiRoot string, h *http.ServeMux) {
	u, err := url.Parse(apiRoot)
	if u.Path == "" || err != nil {
		warn.Panic(err)
	}

	path := "/" + u.String() + "/"
	info.Println("Registering handler for", path+"collections/")
	h.HandleFunc(path+"collections/", requireAcceptTaxii(handleTaxiiCollections))

	info.Println("Registering handler for", path)
	h.HandleFunc(path, requireAcceptTaxii(handleTaxiiAPIRoot))
}

func setupHandler() (*http.ServeMux, error) {
	handler := http.NewServeMux()

	result, err := readResource("taxiiAPIRoots", []interface{}{})
	if err != nil {
		return handler, err
	}
	roots := result.([]string)

	for _, apiRoot := range roots {
		registerAPIRoot(apiRoot, handler)
	}

	handler.HandleFunc("/taxii/", requireAcceptTaxii(handleTaxiiDiscovery))
	handler.HandleFunc("/", requireAcceptTaxii(handleUndefinedRequest))

	return handler, err
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
	server, err := newCabby()
	if err != nil {
		fail.Fatal("Can't start server:", err)
	}
	fail.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
