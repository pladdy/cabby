package main

import (
	"crypto/tls"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	CabbyConfigFile       = "config/cabby.json"
	TAXIIContentType      = "application/vnd.oasis.taxii+json; version=2.0"
	DiscoveryResourceFile = "test_data/discovery.json"
	SixMonthsOfSeconds    = 63072000
)

var (
	info  = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn  = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func basicAuth(h http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		user, pass, ok := r.BasicAuth()

		if !ok || !validate(user, pass) {
			warn.Println("Invalid user/pass combination")
			w.Header().Set("WWW-Authenticate", "Basic realm=TAXII 2.0")
			error := Error{Title: "Unauthorized", HTTPStatus: http.StatusUnauthorized}
			http.Error(w, error.Message(), http.StatusUnauthorized)
			return
		}

		info.Println("Basic Auth validated")
		h(w, r)
	}
}

func validate(u, p string) bool {
	if u == "pladdy" && p == "pants" {
		return true
	}
	return false
}

func strictTransportSecurity() (key, value string) {
	return "Strict-Transport-Security", "max-age=" + strconv.Itoa(SixMonthsOfSeconds) + "; includeSubDomains"
}

func handleDiscovery(w http.ResponseWriter, r *http.Request) {
	w.Header().Add(strictTransportSecurity())

	defer func() {
		if r := recover(); r != nil {
			http.Error(w, "Resource not found", http.StatusNotFound)
		}
	}()

	b := parseDiscoveryResource(DiscoveryResourceFile)
	w.Header().Set("Content-Type", TAXIIContentType)
	info.Println("Handling discovery resource request")
	io.WriteString(w, string(b))
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
	//config := parseConfig(CabbyConfigFile)
	config := CabbyConfig{}.parse(CabbyConfigFile)
	port := strconv.Itoa(config.Port)

	handler := http.NewServeMux()
	handler.HandleFunc("/taxii", basicAuth(handleDiscovery))

	tlsConfig := setupTLS()

	server := &http.Server{
		Addr:         ":" + port,
		Handler:      handler,
		TLSConfig:    tlsConfig,
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}

	info.Println("Serving on port " + port)
	error.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
