package pkg

import log "github.com/sirupsen/logrus"

func SetupLogger() {
	// Set up logrus logging format and level
	log.SetFormatter(&log.JSONFormatter{})
	log.SetLevel(log.InfoLevel)
}
