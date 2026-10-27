package main

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/output"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/ports"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/targets"
)

var version = "dev"

func main() {
	// a) flags
	target := flag.String("target", "127.0.0.1", `hosts, IPs or CIDR ranges, e.g. "scanme.nmap.org,192.168.1.0/24"`)
	portStr := flag.String("ports", "1-1024", `ports, e.g. "22,80,1000-2000"`)
	timeout := flag.Duration("timeout", 3*time.Second, "per-port timeout, e.g. 500ms, 2s")
	workers := flag.Int("workers", 100, "number of ports scanned at once")
	outFile := flag.String("o", "", "also save results to a .json or .csv file")
	all := flag.Bool("all", false, "show closed and filtered ports too")
	flag.Parse()

	// Status messages go to stderr and results to stdout, so
	// "portscanner > results.txt" saves only the table.
	// The art and the progress bar only make sense on a live terminal.
	interactive := isTerminal(os.Stderr)
	if interactive {
		output.Banner(os.Stderr)
	}
	fmt.Fprintln(os.Stderr, "tcp-port-scanner", "v"+version)

	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "error: -workers must be at least 1")
		os.Exit(1)
	}

	if *outFile != "" {
		if err := output.CheckExt(*outFile); err != nil {
			fmt.Fprintln(os.Stderr, "error:", err)
			os.Exit(1)
		}
	}

	// b) ports: "20-25,80" → [20 21 ... 80], or stop
	portList, err := ports.Parse(*portStr)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}

	// c) targets: "scanme.nmap.org,10.0.0.0/30" → names with their IPs, or stop
	entries, err := targets.Parse(*target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	hosts, err := targets.Resolve(entries)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error:", err)
		os.Exit(1)
	}
	ips := make([]string, len(hosts))
	for i, h := range hosts {
		ips[i] = h.IP
	}

	// d) scan and report
	if len(hosts) == 1 {
		fmt.Fprintf(os.Stderr, "Scanning %s (%s), %d ports, %d workers\n\n", hosts[0].Name, hosts[0].IP, len(portList), *workers)
	} else {
		fmt.Fprintf(os.Stderr, "Scanning %d hosts, %d ports each, %d workers\n\n", len(hosts), len(portList), *workers)
	}

	var progress func(done, total int) // nil = no progress bar
	if interactive {
		progress = output.Progress(os.Stderr)
	}

	start := time.Now()
	results := scanner.Scan(ips, portList, *timeout, *workers, progress)
	elapsed := time.Since(start)

	summary := output.Summarize(results, elapsed)
	output.Table(os.Stdout, hosts, results, summary, *all)

	if *outFile != "" {
		rep := output.Report{
			Targets:  hosts,
			Open:     summary.Open,
			Closed:   summary.Closed,
			Filtered: summary.Filtered,
			Elapsed:  elapsed.Round(time.Millisecond).String(),
			Results:  results,
		}
		if err := output.Save(*outFile, rep); err != nil {
			fmt.Fprintln(os.Stderr, "error: cannot save results:", err)
			os.Exit(1)
		}
		fmt.Fprintln(os.Stderr, "saved results to", *outFile)
	}
}

// isTerminal reports whether f is an interactive console rather than a pipe or a file.
func isTerminal(f *os.File) bool {
	fi, err := f.Stat()
	return err == nil && fi.Mode()&os.ModeCharDevice != 0
}
