// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package handlers

import (
	"os"
	"path"
	"path/filepath"

	log "github.com/sirupsen/logrus"
)

// writeFileOnDisk writes the content of the file, in the tmpFolder
func writeFileOnDisk(tmpFolder, filename string, content []byte) error {

	log.WithFields(log.Fields{
		"event":     "writeFile",
		"tmpFolder": tmpFolder,
		"filename":  filename,
	}).Debug("Writing file")

	// if in sub-dir, recreate sub-dir structure
	tDir := path.Dir(filename)
	if tDir != "." && tDir != "/" {
		tpFolder := filepath.Join(tmpFolder, tDir)
		e := os.MkdirAll(tpFolder, 0755)
		if e != nil {
			log.Error(e)
			return e
		}
	}
	fp := filepath.Join(tmpFolder, filepath.Clean(filename))
	tmpfile, err := os.Create(fp)
	if err != nil {
		log.Error(err)
		return err
	}
	defer tmpfile.Close()

	_, err = tmpfile.Write(content)
	if err != nil {
		defer func() {
			e := os.Remove(tmpfile.Name())
			if e != nil {
				log.Error(e)
			}
		}()
		log.Error(err)
		return err
	}

	return nil
}
