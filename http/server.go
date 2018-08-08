package http

import (
	"crypto/tls"
	"net/http"
	"strconv"

	cabby "github.com/pladdy/cabby2"
	log "github.com/sirupsen/logrus"
)

// NewCabby returns a new http server
func NewCabby(ds cabby.DataStore, c cabby.Config) *http.Server {
	handler := http.NewServeMux()

	registerAPIRoots(ds, handler)

	dh := DiscoveryHandler{DiscoveryService: ds.DiscoveryService(), Port: c.Port}
	registerRoute(handler, "taxii", WithAcceptType(RouteRequest(dh), TaxiiContentType))

	registerRoute(handler, "/", handleUndefinedRoute)

	return setupServer(ds, handler, c)
}

func setupServer(ds cabby.DataStore, h http.Handler, c cabby.Config) *http.Server {
	p := strconv.Itoa(c.Port)
	log.WithFields(log.Fields{"port": p}).Info("Server port configured")

	return &http.Server{
		Addr:         ":" + p,
		Handler:      withRequestLogging(withBasicAuth(h, ds.UserService())),
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
