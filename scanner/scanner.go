package scanner

import (
	"github.com/salesforce/lobster-pot/config"
)

// ScanFolder takes a path to a folder to scan, calls the scanning binary to do the scan
// and returns a list of findings, and an error state.
func ScanFolder(tmpFolder string, c config.Config) (findings []Finding, err error) {

	if c.Scanner.Type == "binary" {
		findings, err = scanBinary(tmpFolder, c)
	}

	if c.Scanner.Type == "golang" {
		findings, err = scanEmbeddedGo(tmpFolder, c.Scanner)
	}

	return findings, err
}
