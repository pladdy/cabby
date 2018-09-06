package main

import (
	"flag"
	"os"

	cabby "github.com/pladdy/cabby2"
	"github.com/pladdy/cabby2/sqlite"
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

func main() {
	cabbyEnv := os.Getenv(cabbyEnvironmentVariable)
	if len(cabbyEnv) == 0 {
		cabbyEnv = defaultCabbyEnvironment
	}

	log.WithFields(log.Fields{"environment": cabbyEnv}).Info("Cabby environment set")

	var configPath = flag.String("config", cabbyConfigs[cabbyEnv], "path to cabby config file")
	flag.Parse()

	cs := cabby.Configs{}.Parse(*configPath)
	c := cs[cabbyEnv]

	ds, err := sqlite.NewDataStore(c.DataStore["path"])
	if err != nil {
		log.WithFields(log.Fields{"error": err}).Panic("Can't connect to data store")
	}
}
