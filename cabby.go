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

	handler, err := setupRouteHandler(ts, c.Port)
	if err != nil {
		return &server, err
	}
	return setupServer(ts, handler, c), err
}

// server is set up with basic auth and HSTS applied to each handler
func setupServer(ts taxiiStorer, h http.Handler, c config) *http.Server {
	p := strconv.Itoa(c.Port)
	log.WithFields(log.Fields{"port": p}).Info("Server port configured")

	return &http.Server{
		Addr:         ":" + p,
		Handler:      withRequestLogging(withBasicAuth(h, ts)),
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
	log.WithFields(log.Fields{"environment": cabbyEnv}).Info("Cabby environment set")

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
