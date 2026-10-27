package output

import (
	"fmt"
	"io"
	"text/tabwriter"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/services"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/targets"
)

// Summary holds the totals shown under the table.
type Summary struct {
	Hosts                  int
	Open, Closed, Filtered int
	Elapsed                time.Duration
}

// Summarize counts how many hosts were scanned and how many results are open, closed and filtered.
func Summarize(results []scanner.Result, elapsed time.Duration) Summary {
	s := Summary{Elapsed: elapsed}
	hosts := make(map[string]bool)

	for _, r := range results {
		hosts[r.Host] = true
		switch r.State {
		case scanner.Open:
			s.Open++
		case scanner.Closed:
			s.Closed++
		default:
			s.Filtered++
		}
	}

	s.Hosts = len(hosts)
	return s
}

// Table writes one aligned table per host, showing open ports (or all ports if showAll),
// then the summary. Hosts with nothing to show are skipped.
func Table(w io.Writer, hosts []targets.Target, results []scanner.Result, s Summary, showAll bool) {
	names := make(map[string]string, len(hosts)) // IP → name the user typed
	for _, h := range hosts {
		names[h.IP] = h.Name
	}

	shown := 0
	// results are grouped by host: [i, j) is one host's block
	for i := 0; i < len(results); {
		j := i
		for j < len(results) && results[j].Host == results[i].Host {
			j++
		}
		if hostTable(w, label(results[i].Host, names), results[i:j], showAll) {
			shown++
		}
		i = j
	}

	if shown == 0 {
		fmt.Fprint(w, "No open ports found.\n\n")
	}
	fmt.Fprintf(w, "%d open, %d closed, %d filtered on %d %s in %v\n",
		s.Open, s.Closed, s.Filtered, s.Hosts, plural(s.Hosts, "host"), s.Elapsed.Round(time.Millisecond))
}

// hostTable writes one host's heading and table. It writes nothing and
// returns false if none of the host's results are shown.
func hostTable(w io.Writer, heading string, results []scanner.Result, showAll bool) bool {
	var rows []scanner.Result
	for _, r := range results {
		if r.State == scanner.Open || showAll {
			rows = append(rows, r)
		}
	}
	if len(rows) == 0 {
		return false
	}

	fmt.Fprintln(w, heading)
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	fmt.Fprintln(tw, "PORT\tSTATE\tSERVICE\tBANNER")
	for _, r := range rows {
		fmt.Fprintf(tw, "%d\t%s\t%s\t%s\n", r.Port, r.State, services.Name(r.Port), r.Banner)
	}
	tw.Flush()
	fmt.Fprintln(w)
	return true
}

// label is "name (ip)", or just the IP when the user typed the IP itself.
func label(ip string, names map[string]string) string {
	if name := names[ip]; name != "" && name != ip {
		return fmt.Sprintf("%s (%s)", name, ip)
	}
	return ip
}

// plural returns word, with an "s" unless n is 1.
func plural(n int, word string) string {
	if n == 1 {
		return word
	}
	return word + "s"
}
