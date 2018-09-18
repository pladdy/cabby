package main

import (
	"flag"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/http"
	"github.com/pladdy/cabby2/sqlite"
	log "github.com/sirupsen/logrus"
)

func main() {
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
