package main

import (
	"flag"
	"fmt"
	"net"
	"os"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/ports"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/services"
)

const version = "0.0.1"

func main() {
	fmt.Println("tcp-port-scanner", "v"+version)

	// a) flags
	target := flag.String("target", "127.0.0.1", "host or IP to scan")
	portStr := flag.String("ports", "1-1024", `ports, e.g. "22,80,1000-2000"`)
	timeout := flag.Duration("timeout", 3*time.Second, "per-port timeout, e.g. 500ms, 2s")
	workers := flag.Int("workers", 100, "number of ports scanned at once")
	flag.Parse()

	if *workers < 1 {
		fmt.Fprintln(os.Stderr, "error: -workers must be at least 1")
		os.Exit(1)
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

	open, closed, filtered := 0, 0, 0
	for _, r := range results {
		switch r.State {
		case scanner.Open:
			fmt.Printf("%-7d %-6s %-12s %s\n", r.Port, r.State, services.Name(r.Port), r.Banner)
			open++
		case scanner.Closed:
			closed++
		default:
			filtered++
		}

	}
	fmt.Printf("\n%d open, %d closed, %d filtered in %v\n", open, closed, filtered, elapsed.Round(time.Millisecond))
}
