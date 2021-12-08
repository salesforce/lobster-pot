package config

import (
	"os"

	"github.com/heroku/rollrus"
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

func initLog() {
	var logLevel string
	if logLevel = os.Getenv("LOG_LEVEL"); logLevel == "" {
		logLevel = "info"
	}

	env := os.Getenv("ENVIRON")

	switch logLevel {
	case "trace":
		// Prevent sensitive data from being logged in non local dev.
		if env != "dev" {
			panic("Trace log level is only available in dev environment")
		}
		log.SetLevel(log.TraceLevel)
		// log the calling method, will ease debug.
		log.SetReportCaller(true)
	case "debug":
		log.SetLevel(log.DebugLevel)
		// log the calling method, will ease debug.
		log.SetReportCaller(true)
	case "info":
		log.SetLevel(log.InfoLevel)
	case "warn":
		log.SetLevel(log.WarnLevel)
	case "error":
		log.SetLevel(log.ErrorLevel)
	case "fatal":
		log.SetLevel(log.FatalLevel)
	}

	log.SetOutput(os.Stdout)
	log.Info("Logs configured at log level " + logLevel)

	rt := os.Getenv("ROLLBAR_TOKEN")

	// bail out if no Rollbar token set
	if rt == "" {
		return
	}

	if env == "" {
		env = "unset"
	}

	hook := rollrus.NewHook(rt, env,
		rollrus.WithLevels(log.ErrorLevel, log.PanicLevel, log.FatalLevel),
	)

	log.SetFormatter(&log.TextFormatter{DisableTimestamp: true})
	log.AddHook(hook)

}
