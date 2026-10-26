# Roadmap

Each stage ends with something that **runs** and that you understand completely.
Tick the boxes as you finish them.

## Stage 0: Setup
Go module, git repo, folder layout, a "hello" program. → [write-up](stage-0-setup.md)
- [x] `go mod init` with your GitHub module path
- [x] `git init` and check that `.gitignore` hides `.claude/`
- [x] `go run ./cmd/portscanner` prints something

## Stage 1: Scan ONE port
Connect to `host:port`, decide open / closed / filtered. → [write-up](stage-1-single-port.md)
- [x] `ScanPort(host, port, timeout)` in `internal/scanner`
- [x] Correctly tells **timeout** apart from **refused**
- [x] Tested against an open port, a closed port and a filtered port

**Go you'll learn:** packages, imports, `net.DialTimeout`, `time.Duration`, error handling, `errors.As`, `defer`

## Stage 2: Parse port lists
Turn `"22,80,8000-8010"` into `[]int`, with validation. Your first **unit tests**.
- [ ] `ports.Parse(spec string) ([]int, error)`
- [ ] Rejects garbage: `"abc"`, `"0"`, `"70000"`, `"100-50"`, `"1--5"`
- [ ] Removes duplicates, keeps sorted order
- [ ] Table-driven tests in `parse_test.go`, run with `go test ./...`

**Go you'll learn:** `strings.Split`, `strconv.Atoi`, slices, maps as sets, `testing` package

## Stage 3: A real CLI, scanned sequentially
`-target`, `-ports`, `-timeout` flags; resolve the hostname; loop over the ports.
- [ ] `flag` package with sensible defaults and `-h` help
- [ ] DNS resolution up front (fail fast on a bad hostname)
- [ ] Sequential scan works; **time it**: notice how slow it is

**Go you'll learn:** `flag`, `net.LookupHost`, `os.Exit`, structs (`Result`)

## Stage 4: Concurrency (the big one)
A worker pool of goroutines fed through channels. Same results, dramatically faster.
- [ ] Jobs channel → N workers → results channel
- [ ] `sync.WaitGroup` to know when all workers are done
- [ ] `-workers` flag; results sorted by port before printing
- [ ] Benchmark: sequential vs 100 vs 500 workers (record numbers in this doc)
- [ ] `go test -race ./...` passes

**Go you'll learn:** goroutines, channels, `sync.WaitGroup`, closing channels, data races

## Stage 5: Service names + banner grabbing
- [ ] `services.Name(port)`: 22 → ssh, 80 → http, unknown → `-`
- [ ] For open ports, read the first bytes the service sends (with a read deadline)
- [ ] Clean the banner (strip newlines, cut length, non-printable chars)

**Go you'll learn:** `conn.SetReadDeadline`, `bufio`, byte slices, maps

## Stage 6: Output and polish
- [ ] Aligned table with a summary (`3 open, 990 closed, 31 filtered in 2.4s`)
- [ ] `-open` to show only open ports
- [ ] `-o results.json` / `-o results.csv`
- [ ] README usage section, demo screenshot or GIF, build with `go build`

**Go you'll learn:** `text/tabwriter`, `encoding/json`, `encoding/csv`, struct tags

## Stage 7: Stretch goals (pick any)
- [ ] Progress indicator during long scans
- [ ] Multiple targets / CIDR ranges (`192.168.1.0/24`) for your home lab
- [ ] Rate limiting (`-rate 100` packets/sec): be polite to targets
- [ ] SYN scan with `gopacket` + Npcap (admin rights)
- [ ] GitHub Actions CI running `go vet`, `go test -race`

## Benchmarks (fill in during Stage 4)

| Mode | Target | Ports | Time |
|------|--------|-------|------|
| Sequential | 127.0.0.1 | 1-1024 | |
| 100 workers | 127.0.0.1 | 1-1024 | |
| 500 workers | 127.0.0.1 | 1-1024 | |
