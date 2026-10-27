package output

import (
	"bytes"
	"strings"
	"testing"
	"time"

	"github.com/abhishekshinde2104/tcp-port-scanner/internal/scanner"
	"github.com/abhishekshinde2104/tcp-port-scanner/internal/targets"
)

var (
	testHosts = []targets.Target{
		{Name: "scanme.nmap.org", IP: "45.33.32.156"},
		{Name: "10.0.0.1", IP: "10.0.0.1"},
	}
	testResults = []scanner.Result{
		{Host: "45.33.32.156", Port: 22, State: scanner.Open, Banner: "SSH-2.0-OpenSSH"},
		{Host: "45.33.32.156", Port: 23, State: scanner.Closed},
		{Host: "10.0.0.1", Port: 22, State: scanner.Filtered},
		{Host: "10.0.0.1", Port: 23, State: scanner.Closed},
	}
)

func TestSummarize(t *testing.T) {
	got := Summarize(testResults, time.Second)
	want := Summary{Hosts: 2, Open: 1, Closed: 2, Filtered: 1, Elapsed: time.Second}
	if got != want {
		t.Errorf("Summarize = %+v, want %+v", got, want)
	}
}

func TestTable(t *testing.T) {
	s := Summarize(testResults, 1500*time.Millisecond)

	var buf bytes.Buffer // a bytes.Buffer is an io.Writer, so we can capture the table
	Table(&buf, testHosts, testResults, s, false)
	out := buf.String()

	for _, want := range []string{
		"scanme.nmap.org (45.33.32.156)",
		"22    open   ssh      SSH-2.0-OpenSSH",
		"1 open, 2 closed, 1 filtered on 2 hosts in 1.5s",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	// 10.0.0.1 has no open ports, so without -all it gets no block at all
	if strings.Contains(out, "10.0.0.1") || strings.Contains(out, "closed  ") {
		t.Errorf("non-open rows shown without showAll:\n%s", out)
	}

	buf.Reset()
	Table(&buf, testHosts, testResults, s, true)
	if !strings.Contains(buf.String(), "10.0.0.1\nPORT") {
		t.Errorf("showAll should print a 10.0.0.1 block:\n%s", buf.String())
	}
}

func TestTableNothingOpen(t *testing.T) {
	res := testResults[2:]
	var buf bytes.Buffer
	Table(&buf, testHosts, res, Summarize(res, time.Second), false)
	if !strings.Contains(buf.String(), "No open ports found.") {
		t.Errorf("want a 'No open ports found.' line:\n%s", buf.String())
	}
}

func TestCSV(t *testing.T) {
	res := []scanner.Result{{Host: "10.0.0.1", Port: 21, State: scanner.Open, Banner: "=HYPERLINK(\"x\")"}}
	var buf bytes.Buffer
	if err := CSV(&buf, res); err != nil {
		t.Fatal(err)
	}
	want := "host,port,state,service,banner\n10.0.0.1,21,open,ftp,\"'=HYPERLINK(\"\"x\"\")\"\n"
	if buf.String() != want {
		t.Errorf("CSV =\n%q\nwant\n%q", buf.String(), want)
	}
}

func TestProgress(t *testing.T) {
	var buf bytes.Buffer
	p := Progress(&buf)
	for done := 1; done <= 1000; done++ {
		p(done, 1000)
	}
	// 1000 calls, but only the 101 distinct percentages (0..100) are drawn
	if n := strings.Count(buf.String(), "ports"); n != 101 {
		t.Errorf("drew %d times, want 101 (0%%..100%%)", n)
	}
	if !strings.HasSuffix(buf.String(), "\r") {
		t.Error("progress line should be cleared at the end")
	}
}

func TestCheckExt(t *testing.T) {
	for path, ok := range map[string]bool{
		"r.json": true, "r.CSV": true, "dir/r.csv": true,
		"r.txt": false, "r": false, "json": false,
	} {
		if err := CheckExt(path); (err == nil) != ok {
			t.Errorf("CheckExt(%q) error = %v, want ok=%v", path, err, ok)
		}
	}
}
