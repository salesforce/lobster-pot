package config

import "fmt"

// for multiple orgs, the env vars the variables are in the form  GITHUB_<VARNAME>_<NUMERICAL_ID>
// buildEnvVerName returns the env var name for the var and id
func buildEnvVarName(index string, name string) string {
	return fmt.Sprintf("%s_%s", name, index)
}
