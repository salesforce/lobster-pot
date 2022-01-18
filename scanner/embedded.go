// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package scanner

import (
	"fmt"

	"github.com/salesforce/lobster-pot/config"
	log "github.com/sirupsen/logrus"
)

func scanEmbeddedGo(tmpFolder string, s config.Scanner) (findings []Finding, err error) {

	log.WithFields(log.Fields{
		"scanner": s.Name,
	}).Debug("Scanning")

	switch s.Name {

	case "wraith":
		findings, err = scanWraith(tmpFolder)

	default:
		err = fmt.Errorf("unknown Embedded scanner named: %s", s.Name)

	}

	return findings, err

}
