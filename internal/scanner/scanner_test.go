package scanner

import (
	"net"
	"testing"
	"time"
)

// listen starts a local TCP server that sends greeting to every client,
// and returns its port. The server stops when the test ends.
func listen(t *testing.T, greeting string) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0") // port 0 = let the OS pick a free port
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { ln.Close() })

	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return // listener closed
			}
			c.Write([]byte(greeting))
			c.Close()
		}
	}()
	return ln.Addr().(*net.TCPAddr).Port
}

// closedPort returns a port that nothing is listening on.
func closedPort(t *testing.T) int {
	t.Helper()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := ln.Addr().(*net.TCPAddr).Port
	ln.Close()
	return port
}

func TestScanPort(t *testing.T) {
	tests := []struct {
		name       string
		greeting   string
		wantBanner string
	}{
		{"banner first line only", "SSH-2.0-Test_1.0\r\nsecond line\n", "SSH-2.0-Test_1.0"},
		{"silent service", "", ""},
		{"escape codes are neutralised", "\x1b[2Jevil", ".[2Jevil"},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			port := listen(t, tc.greeting)
			state, banner := ScanPort("127.0.0.1", port, time.Second)
			if state != Open {
				t.Fatalf("state = %q, want %q", state, Open)
			}
			if banner != tc.wantBanner {
				t.Errorf("banner = %q, want %q", banner, tc.wantBanner)
			}
		})
	}
}

func TestScanPortClosed(t *testing.T) {
	state, banner := ScanPort("127.0.0.1", closedPort(t), time.Second)
	if state != Closed || banner != "" {
		t.Errorf("got (%q, %q), want (%q, \"\")", state, banner, Closed)
	}
}

func TestScan(t *testing.T) {
	open := listen(t, "hi\n")
	closed := closedPort(t)
	hosts := []string{"127.0.0.2", "127.0.0.1"} // order given here must be kept
	ports := []int{open, closed}

	var calls, lastDone, lastTotal int
	progress := func(done, total int) {
		calls++
		lastDone, lastTotal = done, total
	}

	res := Scan(hosts, ports, time.Second, 10, progress)

	if len(res) != 4 {
		t.Fatalf("got %d results, want 4", len(res))
	}
	if calls != 4 || lastDone != 4 || lastTotal != 4 {
		t.Errorf("progress: %d calls, last (%d, %d), want 4 calls ending at (4, 4)", calls, lastDone, lastTotal)
	}

	lo, hi := min(open, closed), max(open, closed)
	wantOrder := []struct {
		host string
		port int
	}{{"127.0.0.2", lo}, {"127.0.0.2", hi}, {"127.0.0.1", lo}, {"127.0.0.1", hi}}
	for i, w := range wantOrder {
		if res[i].Host != w.host || res[i].Port != w.port {
			t.Errorf("res[%d] = %s:%d, want %s:%d", i, res[i].Host, res[i].Port, w.host, w.port)
		}
	}

	for _, r := range res {
		if r.Host == "127.0.0.1" && r.Port == open && r.State != Open {
			t.Errorf("127.0.0.1:%d = %q, want open", open, r.State)
		}
	}
}

func TestScanNilProgress(t *testing.T) {
	res := Scan([]string{"127.0.0.1"}, []int{closedPort(t)}, time.Second, 1, nil)
	if len(res) != 1 {
		t.Errorf("got %d results, want 1", len(res))
	}
}
