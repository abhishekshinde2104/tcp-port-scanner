package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/output"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/ports"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
)

var version = "dev"

func main() {
	fmt.Println("tcp-port-scanner", "v"+version)

	// a) flags
	target := flag.String("target", "127.0.0.1", "host or IP to scan")
	portStr := flag.String("ports", "1-1024", `ports, e.g. "22,80,1000-2000"`)
	timeout := flag.Duration("timeout", 3*time.Second, "per-port timeout, e.g. 500ms, 2s")
	workers := flag.Int("workers", 100, "number of ports scanned at once")
	outFile := flag.String("o", "", "also save results to a .json or .csv file")
	all := flag.Bool("all", false, "show closed and filtered ports too")
	flag.Parse()

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

	// c) target: "scanme.nmap.org" → "45.33.32.156", or stop
	addrs, err := net.LookupHost(*target)
	if err != nil {
		fmt.Fprintln(os.Stderr, "error: cannot resolve", *target)
		os.Exit(1)
	}
	ip := addrs[0]

	// d) scan and report
	fmt.Printf("Scanning %s (%s), %d ports, %d workers\n\n", *target, ip, len(portList), *workers)
	start := time.Now()
	results := scanner.Scan(ip, portList, *timeout, *workers)
	elapsed := time.Since(start)

	summary := output.Summarize(results, elapsed)
	output.Table(os.Stdout, results, summary, *all)

	if *outFile != "" {
		rep := output.Report{
			Target:   *target,
			IP:       ip,
			Open:     summary.Open,
			Closed:   summary.Closed,
			Filtered: summary.Filtered,
			Elapsed:  elapsed.Round(time.Millisecond).String(),
			Ports:    results,
		}
		if err := output.Save(*outFile, rep); err != nil {
			fmt.Fprintln(os.Stderr, "error: cannot save results:", err)
			os.Exit(1)
		}
		fmt.Println("saved results to", *outFile)
	}
}
