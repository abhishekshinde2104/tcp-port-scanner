package scanner

import (
	"errors"
	"net"
	"strconv"
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

// Scan checks each port on host one after another and returns the results.
func Scan(host string, ports []int, timeout time.Duration) []Result {
	res := make([]Result, 0, len(ports))

	for _, port := range ports {
		state := ScanPort(host, port, timeout)
		res = append(res, Result{Port: port, State: state})
	}

	return res
}
