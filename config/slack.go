package config

import (
	"fmt"
	"os"
	"regexp"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type SlackApp struct {
	Id            string
	Channel       string
	Token         string
	SigningSecret string
}

type SlackAppID string

type SlackApps map[SlackAppID]SlackApp

func buildSLackAppsConfigs() (apps SlackApps, err error) {
	apps = make(SlackApps)
	re := regexp.MustCompile("^SLACK_APPID_([\\d]+)=(.*)$")
	for _, element := range os.Environ() {

		match := re.FindStringSubmatch(element)
		if len(match) > 0 {
			index := match[1]
			name := SlackAppID(match[2])
			c, e := buildSlackAppConfigFromEnviron(index)
			if e != nil {
				log.WithFields(log.Fields{
					"index": index,
					"name":  name,
				}).Error("Error parsing env vars")
				return nil, e
			}
			apps[name] = *c
		}
	}
	return apps, nil
}

func buildSlackAppConfigFromEnviron(index string) (app *SlackApp, err error) {

	l := log.WithFields(log.Fields{
		"function": "buildSlackAppConfigFromEnviron",
		"index":    index,
	})

	appID := os.Getenv(buildEnvVarName(index, "SLACK_APPID"))
	if appID == "" {
		err = fmt.Errorf("SLACK_APPID not set")
		l.Error(err.Error())
		return nil, err
	}
	channel := os.Getenv(buildEnvVarName(index, "SLACK_CHANNEL"))
	if channel == "" {
		err = fmt.Errorf("SLACK_CHANNEL not set")
		l.Error(err.Error())
		return nil, err
	}
	token := os.Getenv(buildEnvVarName(index, "SLACK_TOKEN"))
	if token == "" {
		err = fmt.Errorf("SLACK_TOKEN not set")
		l.Error(err.Error())
		return nil, err
	}
	signingSecret := os.Getenv(buildEnvVarName(index, "SLACK_SIGNING_SECRET"))
	if signingSecret == "" {
		err = fmt.Errorf("SLACK_SIGNING_SECRET not set")
		l.Error(err.Error())
		return nil, err
	}

	app = &SlackApp{
		Id:            appID,
		Channel:       channel,
		Token:         token,
		SigningSecret: signingSecret,
	}
	return app, nil
}
