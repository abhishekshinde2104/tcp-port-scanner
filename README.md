# tcp-port-scanner

A concurrent TCP connect port scanner written in Go. It is built stage by stage
as a learning project in network programming and security tooling.

> **Status:** 🚧 In development, currently at **Stage 0 (setup)**. See the [roadmap](docs/ROADMAP.md).

## What it does (when finished)

- Scans a host for open TCP ports using a full **TCP connect scan**
- Classifies every port as **open**, **closed** or **filtered**
- Accepts flexible port lists: `80`, `1-1024`, `22,80,443`, `20-25,80,8000-8100`
- Scans hundreds of ports at once with a goroutine **worker pool**
- Labels common services (22 → ssh, 443 → https, ...)
- Grabs service **banners** from open ports (e.g. `SSH-2.0-OpenSSH_8.9`)
- Prints a readable table or machine-readable **JSON / CSV**

## Usage

> Filled in as the CLI is built (Stage 3).

```bash
go run ./cmd/portscanner -target scanme.nmap.org -ports 1-1024
```

## Project layout

```
tcp_port_scanner/
├── cmd/
│   └── portscanner/      # main.go: CLI entry point (flags, wiring)
├── internal/
│   ├── scanner/          # the core: connect to one port, worker pool
│   ├── ports/            # parse "22,80,1000-2000" into a list of ports
│   ├── services/         # port number → service name table
│   └── output/           # print results as a table / JSON / CSV
├── docs/                 # concepts, roadmap and per-stage write-ups
├── go.mod
└── README.md
```

## Learning write-ups

| Doc | Topic |
|-----|-------|
| [00 - Concepts](docs/00-concepts.md) | Ports, the TCP handshake, open/closed/filtered, connect vs SYN scans |
| [Roadmap](docs/ROADMAP.md) | All stages and progress checklist |
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
