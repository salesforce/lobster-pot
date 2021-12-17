package scanner

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"

	"github.com/salesforce/lobster-pot/config"
	log "github.com/sirupsen/logrus"
)

type BinaryScannerOutput struct {
	cmdOutput   []byte
	scannerName string
}

func scanBinary(tmpFolder string, c config.Config) (findings []Finding, err error) {
	b := c.Scanner.Binary

	// Building arguments for the binary, which is a bit ugly.
	// replace %s with the tmpFolder
	argstring := c.Scanner.Arguments
	args := fmt.Sprintf(argstring, tmpFolder)
	a := strings.Split(args, ";")

	cmd := exec.Command(b, a...)
	var out bytes.Buffer
	cmd.Stdout = &out

	log.WithFields(log.Fields{
		"scanner": c.Scanner.Name,
		"type":    c.Scanner.Type,
		"binary":  b,
		"args":    a,
	}).Debug("Running scanner")

	if err := cmd.Run(); err != nil {
		log.Trace(string(out.Bytes()))
		log.Error(err)
		return nil, err
	}

	// Parse output
	s := BinaryScannerOutput{
		cmdOutput:   out.Bytes(),
		scannerName: c.Scanner.Name,
	}

	findings, err = parseFindingsFromBinaryScanner(s)

	return findings, err

}

func parseFindingsFromBinaryScanner(s BinaryScannerOutput) (f []Finding, err error) {

	log.WithFields(log.Fields{
		"scanner": s.scannerName,
	}).Debug("parsing findings")

	log.WithFields(log.Fields{
		"output": (string)(s.cmdOutput),
	}).Trace()

	switch s.scannerName {
	case "grover":
		f, err = ParseGroverFindings(s.cmdOutput)
	case "semgrep":
		f, err = ParseSemgrepFindings(s.cmdOutput)
	default:
		err = fmt.Errorf("unknown binary scanner: %s", s.scannerName)

	}

	return f, err

}
