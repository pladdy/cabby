package http

import (
	"crypto/tls"
	"net/http"
	"strconv"

	"github.com/pladdy/cabby"
	log "github.com/sirupsen/logrus"
)

// NewCabby returns a new http server
func NewCabby(ds cabby.DataStore, c cabby.Config) *http.Server {
	handler := http.NewServeMux()

	registerAPIRoots(ds, handler)

	dh := DiscoveryHandler{DiscoveryService: ds.DiscoveryService(), Port: c.Port}
	registerRoute(handler, "taxii", WithMimeType(routeRequest(dh), "Accept", cabby.TaxiiContentType))

	registerRoute(handler, "/", handleUndefinedRoute)

	return setupServer(ds, handler, c)
}

func setupServer(ds cabby.DataStore, h http.Handler, c cabby.Config) *http.Server {
	p := strconv.Itoa(c.Port)
	log.WithFields(log.Fields{"port": p}).Info("Server port configured")

	return &http.Server{
		Addr:         ":" + p,
		Handler:      withBasicAuth(withLogging(h), ds.UserService()),
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
			// TLS 1.2
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,

			// TLS 1.3
			tls.TLS_AES_256_GCM_SHA384,
			tls.TLS_CHACHA20_POLY1305_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
		},
	}
}
