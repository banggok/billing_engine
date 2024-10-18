package pkg

import (
	"sync"

	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
)

var once sync.Once

// LoadEnv ensures that the environment variables are loaded only once.
func LoadEnv(envFile string) {
	once.Do(func() {
		if err := godotenv.Load(envFile); err != nil {
			log.WithFields(log.Fields{
				"envFile": envFile,
				"error":   err,
			}).Warn("Error loading environment file")
		} else {
			log.WithField("envFile", envFile).Info("Environment variables loaded")
		}
	})
}
