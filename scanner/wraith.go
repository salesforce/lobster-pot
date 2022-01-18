// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package scanner

import (
	wraith "github.com/N0MoreSecr3ts/wraith/core"
)

type wraithFinding struct {
	Action          string `json:"Action"`
	Content         string `json:"Content"`
	CommitAuthor    string `json:"CommitAuthor"`
	CommitHash      string `json:"CommitHash"`
	CommitMessage   string `json:"CommitMessage"`
	CommitURL       string `json:"CommitURL"`
	Description     string `json:"Description"`
	FilePath        string `json:"FilePath"`
	FileURL         string `json:"FileURL"`
	WraithVersion   string `json:"WraithVersion"`
	Hash            string `json:"Hash"`
	LineNumber      string `json:"LineNumber"`
	RepositoryName  string `json:"RepositoryName"`
	RepositoryOwner string `json:"RepositoryOwner"`
	RepositoryURL   string `json:"RepositoryURL"`
	SecretID        string `json:"SecretID"`
	SignatureID     string `json:"SignatureID"`
}

type wraithFindings []wraithFinding

func scanWraith(tmpFolder string) (Findings, error) {

	// TODO: Viper.set some variables if needed
	wraithConfig := wraith.SetConfig()
	scanType := "localPath"
	sess := wraith.NewSession(wraithConfig, scanType)
	sess.Silent = true

	wraith.ScanDir(tmpFolder, sess)

	wf := sess.Findings

	f := make(Findings, len(wf))
	for i, w := range wf {
		nf := Finding{
			FilePath:        w.FilePath,
			LineNumber:      w.LineNumber,
			RuleDescription: w.Description,
			Scanner:         "wraith",
			Secret:          w.Content,
		}
		f[i] = nf
	}

	return f, nil
}
