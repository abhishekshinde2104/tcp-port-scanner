package scanner

import (
	"cmp"
	"errors"
	"net"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

const (
	Open     = "open"
	Closed   = "closed"
	Filtered = "filtered"
)

// Result is the outcome of scanning one port.
type Result struct {
	Port   int    `json:"port"`
	State  string `json:"state"`
	Banner string `json:"banner,omitempty"` // leaves the field out when it's ""
}

// ScanPort tries a TCP connection to host:port and returns its state ("open", "closed" or "filtered") and, if open, the service banner.
func ScanPort(host string, port int, timeout time.Duration) (string, string) {
	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)

	if err == nil {
		banner := grabBanner(conn, timeout)
		conn.Close()
		return Open, banner
	}

	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Timeout() {
		return Filtered, ""
	}

	return Closed, ""
}

// Scan checks ports on host using workers concurrent goroutines and returns the results sorted by port.
func Scan(host string, ports []int, timeout time.Duration, workers int) []Result {
	jobs := make(chan int)       // ports to scan
	results := make(chan Result) // finished results
	var wg sync.WaitGroup

	// WORKERS: take from jobs, put into results
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				state, banner := ScanPort(host, port, timeout)
				results <- Result{Port: port, State: state, Banner: banner}
			}
		}()
	}

	// FEEDER: put every port into jobs, then close jobs
	go func() {
		for _, p := range ports {
			jobs <- p
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	res := make([]Result, 0, len(ports))
	for r := range results {
		res = append(res, r)
	}

	slices.SortFunc(res, func(a, b Result) int {
		return cmp.Compare(a.Port, b.Port)
	})
	return res
}

// grabBanner reads the first line a service sends on connect, cleaned for safe printing.
func grabBanner(conn net.Conn, timeout time.Duration) string {
	conn.SetReadDeadline(time.Now().Add(timeout))

	buf := make([]byte, 256) // a 256-byte buffer
	n, _ := conn.Read(buf)   // n = how many bytes actually arrived (0 if nothing)
	s := string(buf[:n])     // convert only those n bytes to text

	line, _, _ := strings.Cut(s, "\n") // the part before the first newline
	line = strings.TrimSpace(line)     // also strips the trailing \r

	clean := strings.Map(func(r rune) rune {
		if unicode.IsPrint(r) {
			return r
		}
		return '.'
	}, line)

	if r := []rune(clean); len(r) > 60 {
		clean = string(r[:60])
	}
	return clean
}
