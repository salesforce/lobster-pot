// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package scanner

import (
	"encoding/json"
)

type Finding struct {
	FilePath        string `json:"file_path"`
	LineNumber      string `json:"line_number"`
	RuleDescription string `json:"rule_description"`
	Scanner         string `json:"scanner"`
	Secret          string `json:"secret"`
}

type Findings []Finding

func (f Findings) String() string {
	b, err := json.Marshal(f)
	if err != nil {
		return ""
	}
	return string(b)
}
