package output

import (
	"encoding/csv"
	"io"
	"strconv"
	"strings"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/services"
)

// CSV writes results to w as CSV with a header row.
func CSV(w io.Writer, results []scanner.Result) error {
	cw := csv.NewWriter(w)
	cw.Write([]string{"host", "port", "state", "service", "banner"})

	for _, r := range results {
		cw.Write([]string{
			r.Host,
			strconv.Itoa(r.Port),
			r.State,
			services.Name(r.Port),
			safeCell(r.Banner),
		})
	}

	cw.Flush()
	return cw.Error()
}

// safeCell stops spreadsheet apps from running a cell as a formula.
func safeCell(s string) string {
	if s != "" && strings.ContainsAny(s[:1], "=+-@") {
		return "'" + s
	}
	return s
}
