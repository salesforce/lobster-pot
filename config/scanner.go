package config

import (
	"fmt"
	"os"

	_ "github.com/joho/godotenv/autoload"
)

type Scanner struct {
	Type      string // Can be "binary" or "embedded"
	Name      string // Scanner name is used to identify the scanner and unmarshal the output
	RulesFile string
	Binary    string
	Arguments string // Arguments to pass to the scanner, %s will be replaced with the target folder
}

func buildScannerConfig() (s Scanner, err error) {

	// TODO: allow for multiple scanners at the same time

	stype := os.Getenv("SCANNER_TYPE")

	if stype == "" {
		stype = "binary"
	}

	if !isValidScannerType(stype) {
		err = fmt.Errorf("Invalid scanner type: %s", stype)
		return Scanner{}, err
	}

	s.Type = stype

	if s.Type == "binary" {
		binary := os.Getenv("SCANNER_BINARY")
		if binary == "" {
			err = fmt.Errorf("No scanner binary specified")
			return Scanner{}, err
		}
		s.Binary = binary

		// The arguments passed to the binary, each separated by a semi-colon.
		arguments := os.Getenv("SCANNER_ARGUMENTS")
		if arguments == "" {
			err = fmt.Errorf("No scanner arguments specified")
			return Scanner{}, err
		}
		s.Arguments = arguments
	}

	// Scanner Name can be grover, semgrep, or wraith right now. Will add more later.
	sname := os.Getenv("SCANNER_NAME")
	if !IsValidScannerName(sname) {
		err = fmt.Errorf("Invalid scanner name: %s", sname)
		return Scanner{}, err
	}
	s.Name = sname

	return s, nil
}

func IsValidScannerName(name string) bool {
	switch name {
	case
		"wraith",
		"grover",
		"semgrep":
		return true
	}
	return false
}

func isValidScannerType(stype string) bool {
	switch stype {
	case "binary", "golang":
		return true
	}
	return false
}
