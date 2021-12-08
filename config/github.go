package config

import (
	"fmt"
	"os"
	"regexp"
	"strconv"

	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type GithubApp struct {
	ID         int64
	OrgName    string
	Secret     string
	PrivateKey []byte
	InstallID  int64
	SlackAppID SlackAppID
}

type GithubOrgName string

// List of github apps, and their corresponding configs, accessible by the org name
type GithubApps map[GithubOrgName]GithubApp

func buildGithubOrgsConfigs() (apps GithubApps, err error) {
	apps = make(GithubApps)
	log.Trace()
	re := regexp.MustCompile("^GITHUB_ORG_([\\d]+)=(.*)$")
	for _, element := range os.Environ() {

		match := re.FindStringSubmatch(element)
		if len(match) > 0 {
			index := match[1]
			name := GithubOrgName(match[2])
			c, e := buildGHAppConfigFromEnviron(index)
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

// buildGHAppConfigFromEnviron builds a GithubAppConfig from the environment variables
// for multiple orgs, the variables are in the form with GITHUB_<VARNAME>_<NUMERICAL_ID>
func buildGHAppConfigFromEnviron(index string) (ghAppConfig *GithubApp, err error) {
	log.Trace()
	l := log.WithFields(log.Fields{
		"function": "buildGHAppConfigFromEnviron",
		"index":    index,
	})

	orgName := os.Getenv(buildEnvVarName(index, "GITHUB_ORG"))
	if orgName == "" {
		err = fmt.Errorf("GITHUB_ORG not set")
		l.Error(err.Error())
		return nil, err
	}
	secret := os.Getenv(buildEnvVarName(index, "GITHUB_SECRET"))
	if !(os.Getenv("ENVIRON") == "dev") {
		if secret == "" {
			err = fmt.Errorf("GITHUB_SECRET not set")
			l.Error(err.Error())
			return nil, err
		}
	}
	keyData := os.Getenv(buildEnvVarName(index, "GITHUB_PRIVATE_KEY"))
	if keyData == "" {
		err = fmt.Errorf("GITHUB_PRIVATE_KEY not set")
		l.Error(err.Error())
		return nil, err
	}
	eid := os.Getenv(buildEnvVarName(index, "GITHUB_APPID"))
	if eid == "" {
		err = fmt.Errorf("GITHUB_APPID not set")
		l.Error(err.Error())
		return nil, err
	}
	eiid := os.Getenv(buildEnvVarName(index, "GITHUB_INSTALLID"))
	if eiid == "" {
		err = fmt.Errorf("GITHUB_INSTALLID not set")
		l.Error(err.Error())
		return nil, err
	}
	aid, e := strconv.ParseInt(eid, 10, 64)
	if e != nil {
		l.Error("Count not convert gitHubAppID to int ", e)
		return nil, e
	}
	iid, e := strconv.ParseInt(eiid, 10, 64)
	if e != nil {
		l.Error("Count not convert gitHubInstallID to int ", e)
		return nil, e
	}
	slack := os.Getenv(buildEnvVarName(index, "GITHUB_SLACK_APPID"))
	if slack == "" {
		err = fmt.Errorf("GITHUB_SLACK_APPID not set")
		l.Error(err.Error())
		return nil, err

	}

	ghAppConfig = &GithubApp{
		ID:         aid,
		InstallID:  iid,
		PrivateKey: []byte(keyData),
		OrgName:    orgName,
		Secret:     secret,
		SlackAppID: SlackAppID(slack),
	}
	return ghAppConfig, nil
}
