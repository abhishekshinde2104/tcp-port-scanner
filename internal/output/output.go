package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/services"
)

// Summary holds the totals shown under the table.
type Summary struct {
	Open, Closed, Filtered int
	Elapsed                time.Duration
}

// Summarize counts how many results are open, closed and filtered.
func Summarize(results []scanner.Result, elapsed time.Duration) Summary {
	s := Summary{Elapsed: elapsed}

	for _, r := range results {
		switch r.State {
		case scanner.Open:
			s.Open++
		case scanner.Closed:
			s.Closed++
		default:
			s.Filtered++
		}
	}

	return s
}

// Table writes open ports (or all ports if showAll) as an aligned table, then the summary.
func Table(w io.Writer, results []scanner.Result, s Summary, showAll bool) {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSTATE\tSERVICE\tBANNER")

	for _, r := range results {
		if r.State != scanner.Open && !showAll {
			continue
		}
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", r.Port, r.State, services.Name(r.Port), r.Banner)
	}

	tw.Flush()
	fmt.Fprintf(w, "\n%d open, %d closed, %d filtered in %v\n",
		s.Open, s.Closed, s.Filtered, s.Elapsed.Round(time.Millisecond))
}
