# tcp-port-scanner

A fast, concurrent TCP connect port scanner written in Go, with no dependencies outside the standard library.
It was built stage by stage as a learning project in network programming and security tooling.

![demo](docs/demo.gif)

## Features

- **TCP connect scan**: classifies every port as **open**, **closed** or **filtered**
- **Multiple targets and CIDR ranges**: `scanme.nmap.org,10.0.0.5,192.168.1.0/24`
- Flexible port lists: `80`, `1-1024`, `22,80,443`, `20-25,80,8000-8100`
- **Worker pool** of goroutines shared across all hosts: 100 ports to a remote host in ~3 s instead of minutes
- Live **progress bar** while scanning
- Labels common services (22 → ssh, 443 → https, ...)
- Grabs service **banners** from open ports (e.g. `SSH-2.0-OpenSSH_6.6.1p1`)
- Aligned terminal table, plus optional **JSON** or **CSV** output files
- Treats network data as untrusted: banners are sanitised and CSV cells are protected against formula injection

## Install

Build from source (Go 1.25+):

```bash
go build -trimpath -ldflags "-s -w -X main.version=1.0.0" -o bin/portscanner ./cmd/portscanner
```

Or download a prebuilt binary from the Releases page.

## Usage

```
portscanner [flags]
```

| Flag | Default | Description |
|------|---------|-------------|
| `-target` | `127.0.0.1` | Comma-separated host names, IPs and CIDR ranges, e.g. `scanme.nmap.org,192.168.1.0/24` (at most 4096 hosts) |
| `-ports` | `1-1024` | Ports to scan, e.g. `22,80,443` or `1-1024` or `20-25,8080` |
| `-timeout` | `3s` | Per-port connect timeout, e.g. `500ms`, `2s` |
| `-workers` | `100` | Number of ports scanned at the same time |
| `-all` | `false` | Also show closed and filtered ports in the table |
| `-o` | | Also save results to a file; `.json` or `.csv` picks the format |

### Examples

Scan the default range on your own machine:

```bash
portscanner
```

Scan two hosts and save JSON:

```bash
portscanner -target scanme.nmap.org,127.0.0.1 -ports 22,80,135 -o results.json
```

```
scanme.nmap.org (45.33.32.156)
PORT  STATE  SERVICE  BANNER
22    open   ssh      SSH-2.0-OpenSSH_6.6.1p1 Ubuntu-2ubuntu2.13
80    open   http

127.0.0.1
PORT  STATE  SERVICE  BANNER
135   open   msrpc

3 open, 2 closed, 1 filtered on 2 hosts in 3.17s
```

Sweep your own home network for a few common ports (keep the port list short for ranges):

```bash
portscanner -target 192.168.1.0/24 -ports 22,80,443,3389 -timeout 1s -workers 500
```

Show every port, not just open ones, and save CSV:

```bash
portscanner -ports 130-140 -all -o results.csv
```

### Terminal vs pipes

The banner, version line, progress bar and "saved results" messages go to **stderr**. The tables go to **stdout**.
So `portscanner > scan.txt` saves only the results. The ASCII art and the progress bar only appear when stderr
is a live terminal, so they never end up in log files.

### Output files

The terminal shows open ports (or all of them with `-all`). An output file always contains **every** scanned host and port.

**JSON** (shortened)
```json
{
  "targets": [
    { "name": "scanme.nmap.org", "ip": "45.33.32.156" },
    { "name": "127.0.0.1", "ip": "127.0.0.1" }
  ],
  "open": 3,
  "closed": 2,
  "filtered": 1,
  "elapsed": "3.17s",
  "results": [
    { "host": "45.33.32.156", "port": 22, "state": "open", "banner": "SSH-2.0-OpenSSH_6.6.1p1 Ubuntu-2ubuntu2.13" },
    { "host": "45.33.32.156", "port": 135, "state": "filtered" }
  ]
}
```

**CSV**
```
host,port,state,service,banner
45.33.32.156,22,open,ssh,SSH-2.0-OpenSSH_6.6.1p1 Ubuntu-2ubuntu2.13
45.33.32.156,135,filtered,msrpc,
```

An unsupported extension (for example `-o results.txt`) is rejected **before** the scan starts.

## How it works

1. **Parse** the port list into a sorted, de-duplicated `[]int`. Bad input such as `abc`, `0`, `70000` or `100-50` is rejected.
2. **Expand and resolve targets**: CIDR ranges become single addresses (`net/netip`; network and broadcast
   addresses are skipped), each name is resolved once with DNS, and duplicate IPs are dropped. A bad host name fails fast.
3. **Scan** with a worker pool: a feeder goroutine puts every `host:port` pair on a `jobs` channel, N workers call
   `net.DialTimeout` and send each `Result` on a `results` channel, and the main goroutine collects them,
   updates the progress bar, and sorts them by host, then by port.
4. **Classify** each port:
   - connection succeeded → **open**
   - connection refused (RST) → **closed**
   - timeout (no reply) → **filtered**
5. **Banner**: for open ports, read the first line the service sends within the timeout.
6. **Report**: print the table, and optionally write JSON or CSV.

### Notes

- **Windows and closed remote ports:** Windows retries a connection after an RST, so a closed port on a *remote* host
  takes about 2–3 s to report. With a timeout below that, closed ports are misreported as filtered. That is why
  the default timeout is 3 s. Thanks to the worker pool, the cost is about one timeout per scan, not one per port.
- **Banners:** services such as SSH, FTP and SMTP speak first. HTTP waits for the client, so it shows no banner.
- **Scanning ranges:** there is no ping-based host discovery, so a host that is down costs a full timeout per port.
  For a `/24`, use a short port list, a shorter `-timeout`, and more `-workers`.
- **Connect scan vs SYN scan:** this tool completes the full TCP handshake and needs no admin rights.
  A SYN ("half-open") scan needs raw sockets and is listed as a stretch goal.

## Benchmarks

| Mode | Target | Ports | Time |
|------|--------|-------|------|
| Sequential, 3 s timeout | scanme.nmap.org | 20-25 | 14.6 s |
| 100 workers, 3 s timeout | scanme.nmap.org | 1-100 | **2.97 s** |
| 100 workers | 127.0.0.1 | 1-1024 | 107 ms |

More numbers are in the [roadmap](docs/ROADMAP.md#benchmarks-fill-in-during-stage-4).

## Project layout

```
tcp_port_scanner/
├── cmd/
│   └── portscanner/      # main.go: CLI flags and wiring
├── internal/
│   ├── scanner/          # connect to one port, banner grab, worker pool (+ tests with a local fake server)
│   ├── targets/          # parse hosts and CIDR ranges, resolve them (+ tests)
│   ├── ports/            # parse "22,80,1000-2000" into a list of ports (+ tests)
│   ├── services/         # port number → service name (+ tests)
│   └── output/           # tables, JSON, CSV, progress bar, ASCII banner (+ tests)
├── docs/                 # concepts, roadmap and per-stage write-ups
├── go.mod
└── README.md
```

## Tests

```bash
go test ./...
```

The scanner tests start a throwaway TCP server on `127.0.0.1`, so they need no network access and never touch other hosts.

## Learning write-ups

| Doc | Topic |
|-----|-------|
| [00 - Concepts](docs/00-concepts.md) | Ports, the TCP handshake, open/closed/filtered, connect vs SYN scans, concurrency in Go |
| [Roadmap](docs/ROADMAP.md) | All stages, progress checklist and benchmarks |
| [Stage 0 - Setup](docs/stage-0-setup.md) | Go modules, git, project layout |
| [Stage 1 - Scan one port](docs/stage-1-single-port.md) | `net.DialTimeout` and classifying errors |

## ⚠️ Legal and ethical use

Only scan systems you **own** or have **written permission** to test.
Unauthorized scanning may break computer misuse laws and service terms.
Safe targets for practice:

- your own machine: `127.0.0.1`
- Nmap's official test host: `scanme.nmap.org` (scan it lightly, a few times a day at most)

If you do bug bounty work, check each program's policy first. Many programs restrict automated scanning.

## Acknowledgements

The project idea and feature set are inspired by the C++ / Boost.Asio
[simple-port-scanner](https://github.com/CarterPerez-dev/Cybersecurity-Projects/tree/main/PROJECTS/beginner/simple-port-scanner)
from the *Cybersecurity-Projects* collection. This is an independent Go implementation.
