package output

import (
	"encoding/json"
	"io"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/targets"
)

// Report is everything saved to a results file.
type Report struct {
	Targets  []targets.Target `json:"targets"`
	Open     int              `json:"open"`
	Closed   int              `json:"closed"`
	Filtered int              `json:"filtered"`
	Elapsed  string           `json:"elapsed"`
	Results  []scanner.Result `json:"results"`
}

// JSON writes rep to w as indented JSON.
func JSON(w io.Writer, rep Report) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(rep)
}
