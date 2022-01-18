// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package config

import (
	"os"

	"github.com/salesforce/lobster-pot/db"
	log "github.com/sirupsen/logrus"

	_ "github.com/joho/godotenv/autoload"
)

func initDB() (err error) {
	//setup the database if DATABASE_URL is set (postgres for now)
	if dburl := os.Getenv("DATABASE_URL"); dburl != "" {
		if err := db.Connect(dburl); err != nil {
			log.Error(err)
			return err
		}
	}
	return nil
}
