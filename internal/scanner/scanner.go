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
