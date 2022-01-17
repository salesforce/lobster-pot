// Copyright (c) 2022, salesforce.com, inc.
// All rights reserved.
// SPDX-License-Identifier: BSD-3-Clause
// For full license text, see the LICENSE file in the repo root or https://opensource.org/licenses/BSD-3-Clause
package config

import "fmt"

// for multiple orgs, the env vars the variables are in the form  GITHUB_<VARNAME>_<NUMERICAL_ID>
// buildEnvVerName returns the env var name for the var and id
func buildEnvVarName(index string, name string) string {
	return fmt.Sprintf("%s_%s", name, index)
}
