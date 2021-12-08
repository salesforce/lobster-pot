package scanner

import (
	"encoding/json"
	"fmt"
)

type semgrepFindings struct {
	Errors  []semgrepError
	Results []semgrepResult
}

type semgrepError struct{}

type semgrepResults []semgrepResult

type semgrepResult struct {
	CheckID string          `json:"check_id"`
	End     semgrepPosition `json:"end"`
	Extra   semgrepExtra    `json:"extra"`
	Path    string          `json:"path"`
	Start   semgrepPosition `json:"start"`
}

type semgrepPosition struct {
	Column int `json:"col"`
	Line   int `json:"line"`
	Offset int `json:"offset,omitempty"`
}

type semgrepExtra struct {
	IsIgnored bool            `json:"is_ignored"`
	Lines     string          `json:"lines"`
	Message   string          `json:"message"`
	Metadata  semgrepMetadata `json:"metadata"`
	Metavars  semgrepMetavars `json:"metavars"`
	Severity  string          `json:"severity"`
}

type semgrepMetadata struct {
	Category      string   `json:"category"`
	License       string   `json:"license"`
	Source        string   `json:"source"`
	SourceRuleUrl string   `json:"source-rule-url"`
	Technology    []string `json:"technology"`
}

type semgrepMetavars struct{}

func ParseSemgrepFindings(data []byte) (Findings, error) {
	var sf semgrepFindings
	err := json.Unmarshal(data, &sf)
	if err != nil {
		return nil, err
	}

	r := sf.Results

	if len(r) == 0 {
		return Findings{}, nil
	}

	f := make(Findings, len(r))
	for i, fnd := range r {

		// proper formatting of the line number in case of multiline findings
		ln := fmt.Sprintf("%d", fnd.Start.Line)
		if fnd.Start.Line != fnd.End.Line {
			ln = fmt.Sprintf("%d-%d", fnd.Start.Line, fnd.End.Line)
		}
		nf := Finding{
			FilePath:        fnd.Path,
			LineNumber:      ln,
			RuleDescription: fnd.Extra.Message,
			Scanner:         "semgrep",
			Secret:          fnd.Extra.Lines,
		}
		f[i] = nf
	}

	return f, err
}
