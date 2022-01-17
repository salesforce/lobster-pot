// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package scanner

import "encoding/json"

type groverFinding struct {
	Action          string `json:"Action"`
	Comment         string `json:"Comment"`
	CommitAuthor    string `json:"CommitAuthor"`
	CommitHash      string `json:"CommitHash"`
	CommitMessage   string `json:"CommitMessage"`
	CommitURL       string `json:"CommitURL"`
	Description     string `json:"Description"`
	FilePath        string `json:"FilePath"`
	FileURL         string `json:"FileURL"`
	GroverVersion   string `json:"GroverVersion"`
	Hash            string `json:"Hash"`
	LineNumber      string `json:"LineNumber"`
	RepositoryName  string `json:"RepositoryName"`
	RepositoryOwner string `json:"RepositoryOwner"`
	RepositoryURL   string `json:"RepositoryURL"`
	Ruleid          string `json:"Ruleid"`
	RulesVersion    string `json:"RulesVersion"`
	SecretID        string `json:"SecretID"`
	TriagedAs       string `json:"TriagedAs"`
}

type groverFindings []groverFinding

func ParseGroverFindings(data []byte) (Findings, error) {
	var gf groverFindings
	err := json.Unmarshal(data, &gf)
	if err != nil {
		return nil, err
	}

	f := make(Findings, len(gf))
	for i, g := range gf {
		nf := Finding{
			FilePath:        g.FilePath,
			LineNumber:      g.LineNumber,
			RuleDescription: g.Description,
			Scanner:         "grover",
			Secret:          g.Comment,
		}
		f[i] = nf
	}

	return f, err
}
