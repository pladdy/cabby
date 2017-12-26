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
	logInfo  = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	logWarn  = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	logError = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

func registerAPIRoot(apiRoot string, h *http.ServeMux) {
	u, err := url.Parse(apiRoot)
	if u.Path == "" || err != nil {
		logWarn.Panic(err)
	}

	logInfo.Println("Registering API handler for", u)
	h.HandleFunc(u.Path, basicAuth(handleAPIRoot))
}

func setupHandler() *http.ServeMux {
	config := cabbyConfig{}.parse(configPath)
	handler := http.NewServeMux()

	handler.HandleFunc("/taxii", basicAuth(handleDiscovery))

	for _, apiRoot := range config.Discovery.APIRoots {
		if config.validAPIRoot(apiRoot) {
			registerAPIRoot(apiRoot, handler)
		}
	}
	
	handler.HandleFunc("/", handleUndefinedRequest)

	return handler
}

func setupServer(c cabbyConfig, h http.Handler) *http.Server {
	port := strconv.Itoa(c.Port)
	logInfo.Println("Server will listen on port " + port)

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
	config := cabbyConfig{}.parse(configPath)
	handler := setupHandler()
	server := setupServer(config, handler)
	logError.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
