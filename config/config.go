// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package config

import (
	_ "github.com/joho/godotenv/autoload"
	log "github.com/sirupsen/logrus"
)

type Config struct {
	GithubApps GithubApps
	SlackApps  SlackApps
	Scanner    Scanner
}

func Init() (err error) {

	initLog()

	err = initDB()
	if err != nil {
		return err
	}

	return nil
}

func BuildAppsConfig() (Config, error) {
	log.Trace("Building Apps Config")
	gh, e := buildGithubOrgsConfigs()
	if e != nil {
		return Config{}, e
	}

	sl, e := buildSLackAppsConfigs()
	if e != nil {
		return Config{}, e
	}

	sc, e := buildScannerConfig()
	if e != nil {
		return Config{}, e
	}

	return Config{
		GithubApps: gh,
		SlackApps:  sl,
		Scanner:    sc,
	}, nil
}
