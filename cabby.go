package main

import (
	"crypto/tls"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
)

const (
	ConfigPath         = "config/cabby.json"
	TAXIIContentType   = "application/vnd.oasis.taxii+json; version=2.0"
	SixMonthsOfSeconds = 63072000
)

var (
	info  = log.New(os.Stderr, "INFO: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	warn  = log.New(os.Stderr, "WARN: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
	error = log.New(os.Stderr, "ERROR: ", log.Ldate|log.Ltime|log.Lmicroseconds|log.Lshortfile|log.LUTC)
)

/* auth functions */

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

/* resource handlers */

func strictTransportSecurity() (key, value string) {
	return "Strict-Transport-Security", "max-age=" + strconv.Itoa(SixMonthsOfSeconds) + "; includeSubDomains"
}

func resourceNotFound(w http.ResponseWriter) {
	if r := recover(); r != nil {
		http.Error(w, "Resource not found", http.StatusNotFound)
	}
}

func verifyDiscoveryDefined(d DiscoveryResource) {
	if d.Title == "" {
		warn.Panic("Discovery Resource not defined")
	}
}

func handleDiscovery(w http.ResponseWriter, r *http.Request) {
	defer resourceNotFound(w)
	info.Println("Discovery resource requested")

	config := Config{}.parse(ConfigPath)
	verifyDiscoveryDefined(config.Discovery)

	b, err := json.Marshal(config.Discovery)
	if err != nil {
		warn.Panic("Can't serve Discovery:", err)
	}

	w.Header().Set("Content-Type", TAXIIContentType)
	info.Println("Responding with a Discovery resource")
	io.WriteString(w, string(b))
}

/* create server and launch */

// handler wrapper to accept HTTPS requests only
func HSTS(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Add(strictTransportSecurity())
		h.ServeHTTP(w, r)
	})
}

func registerAPIRoots(h http.Handler) {

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

func setupServer(c Config, h http.Handler) *http.Server {
	port := strconv.Itoa(c.Port)
	info.Println("Server will listen on port " + port)

	return &http.Server{
		Addr:         ":" + port,
		Handler:      HSTS(h),
		TLSConfig:    setupTLS(),
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
}

func main() {
	config := Config{}.parse(ConfigPath)

	handler := http.NewServeMux()
	handler.HandleFunc("/taxii", basicAuth(handleDiscovery))

	server := setupServer(config, handler)
	error.Fatal(server.ListenAndServeTLS(config.SSLCert, config.SSLKey))
}
