package scanner

import (
	"cmp"
	"errors"
	"net"
	"slices"
	"strconv"
	"sync"
	"time"
)

const (
	Open     = "open"
	Closed   = "closed"
	Filtered = "filtered"
)

// Result is the outcome of scanning one port.
type Result struct {
	Port  int
	State string
}

// ScanPort tries a TCP connection to host:port and reports "open", "closed" or "filtered"
func ScanPort(host string, port int, timeout time.Duration) string {

	addr := net.JoinHostPort(host, strconv.Itoa(port))
	conn, err := net.DialTimeout("tcp", addr, timeout)

	if err == nil {
		conn.Close()
		return Open
	}

	var netErr net.Error

	if errors.As(err, &netErr) && netErr.Timeout() {
		return Filtered
	}

	return Closed

}

// Scan checks ports on host using workers concurrent goroutines and returns the results sorted by port.
func Scan(host string, ports []int, timeout time.Duration, workers int) []Result {

	jobs := make(chan int)       // ports to scan
	results := make(chan Result) //finished results
	var wg sync.WaitGroup

	// WORKERS: take from jobs, put into results
	for i := 0; i < workers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for port := range jobs {
				state := ScanPort(host, port, timeout)
				results <- Result{Port: port, State: state}
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
