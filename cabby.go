package main

import (
	"crypto/tls"
	"log"
	"net/http"
	"os"
	"strconv"
)

const configPath = "config/cabby.json"

var (
	info = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	err  = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func setupHandler() *http.ServeMux {
	handler := http.NewServeMux()

	handler.HandleFunc("/taxii", basicAuth(handleDiscovery))
	registerAPIRoots(handler)
	//handler.HandleFunc("/admin/collections", basicAuth(adminCollections))
	handler.HandleFunc("/", handleUndefinedRequest)

	return handler
}

func setupServer(c config, h http.Handler) *http.Server {
	port := strconv.Itoa(c.Port)
	info.Println("Server will listen on port " + port)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      hsts(h),
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
	config := config{}.parse(configPath)
	handler := setupHandler()
	server := setupServer(config, handler)
	err.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
