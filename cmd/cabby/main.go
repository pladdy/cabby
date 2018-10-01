package main

import (
	"flag"

	"github.com/pladdy/cabby"
	"github.com/pladdy/cabby/http"
	"github.com/pladdy/cabby/sqlite"
	log "github.com/sirupsen/logrus"
)

func main() {
	log.SetLevel(log.InfoLevel)

	var configPath = flag.String("config", cabby.DefaultProductionConfig, "path to cabby config file")
	flag.Parse()

	c := cabby.Config{}.Parse(*configPath)

	ds, err := sqlite.NewDataStore(c.DataStore["path"])
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("Can't start server")
	}

	server := http.NewCabby(ds, c)
	log.Fatal(server.ListenAndServeTLS(c.SSLCert, c.SSLKey))
}
