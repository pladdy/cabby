package main

import (
	"flag"
	"os"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/http"
	"github.com/pladdy/cabby2/sqlite"
	log "github.com/sirupsen/logrus"
)

func main() {
	cabbyEnv := os.Getenv(cabby.CabbyEnvironmentVariable)
	if len(cabbyEnv) == 0 {
		cabbyEnv = cabby.DefaultCabbyEnvironment
	}

	log.WithFields(log.Fields{"environment": cabbyEnv}).Info("Cabby environment set")

	var configPath = flag.String("config", cabby.CabbyConfigs[cabbyEnv], "path to cabby config file")
	flag.Parse()

	cs := cabby.Configs{}.Parse(*configPath)
	c := cs[cabbyEnv]

	ds, err := sqlite.NewDataStore(c.DataStore["path"])
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("Can't start server")
	}

	server := http.NewCabby(ds, c)
	log.Fatal(server.ListenAndServeTLS(c.SSLCert, c.SSLKey))
}
