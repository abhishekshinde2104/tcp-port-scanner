# 00 - Concepts

The networking ideas you need before writing any code.

## Hosts and ports

A **host** is any device with an IP address: a server, your laptop, a phone, a
router, a printer, a VM or a container. "Server" is a *role*: any program that
waits for connections on a port is acting as a server.

A **port** is a number from 1 to 65535 that says *which program* on the host
should receive the traffic. If the IP address is a building's street address,
the ports are its apartment numbers.

- Every IP address has **65,535 TCP ports** (port 0 is reserved) and a separate set of 65,535 UDP ports.
- Many services can run on one host at once: SSH on 22, HTTPS on 443, MySQL on 3306...
- Most ports are closed. Only ports with a program **listening** on them accept connections.

Try it on your machine:

```powershell
Get-NetTCPConnection -State Listen | Sort-Object LocalPort | Select-Object LocalAddress, LocalPort, OwningProcess
```

## The TCP three-way handshake

```
Scanner                       Target
   | ---- SYN --------------->  |   "Can we talk on port 80?"
   | <--- SYN-ACK ------------  |   "Yes, I'm listening"      → OPEN
   | ---- ACK --------------->  |   "Connected"
```

If nothing listens on the port, the target answers the SYN with **RST** ("go away").

## The three port states

| State | What happened | How our Go code sees it |
|-------|---------------|-------------------------|
| **open** | Target answered SYN-ACK; the connection succeeded | `net.DialTimeout` returns **no error** |
| **closed** | Target answered RST | dial returns an error **quickly** ("connection refused") |
| **filtered** | No answer at all; a firewall probably dropped the packet | dial returns a **timeout** error |

"Filtered" is a guess. A timeout can also mean the network is just slow or the
host is down. That's why the timeout value matters.

## Connect scan vs SYN scan

| | Connect scan (what we build) | SYN "half-open" scan (stretch goal) |
|---|---|---|
| Packets | SYN → SYN-ACK → ACK, then close | SYN → SYN-ACK → **RST** (never finish) |
| How | Normal OS socket (`net.Dial`) | **Raw socket**: you build the packet yourself |
| Privileges | None | Admin/root **on the scanning machine** |
| Windows | Works out of the box | Needs the Npcap driver |
| Logs | The service may log the connection | The service never sees a connection; firewalls/IDS still do |

**Raw socket:** a normal socket lets the OS build the packets and run the handshake for
you. A raw socket hands you the packet bytes and flags directly. That power also allows
spoofing, so operating systems restrict it to admins.

## Concurrent vs parallel vs asynchronous

These words are related, but they don't mean the same thing. Picture a restaurant kitchen:

| Word | Meaning | Kitchen | Describes |
|------|---------|---------|-----------|
| **Concurrent** | several tasks *in progress* during the same period | one chef with pasta boiling, sauce simmering and bread in the oven, switching between them | how work is **structured** |
| **Parallel** | several tasks running at the *exact same instant* | three chefs, each cooking one dish at once (more CPU cores) | how work is **executed** |
| **Asynchronous** | start an operation, don't wait, get notified when it's done | put bread in the oven, set a timer, do other work until it beeps | one **way of calling** an operation (the opposite is *synchronous*: standing at the oven until it's done) |

Key relationships:

- **Asynchronous is one technique for concurrency.** Threads and goroutines are another.
- **Concurrency doesn't need parallelism.** One CPU core can juggle thousands of concurrent tasks by switching between them.

## Why speed comes from concurrency

A scanner spends almost all its time **waiting** for a SYN-ACK, an RST or a timeout.
The CPU does very little.

- One port at a time with a 1 s timeout → 1,000 ports can take about **16 minutes**.
- 500 ports waiting at once → a few **seconds**.

So **concurrency** is what makes a scanner fast. **Parallelism** barely matters, because the CPU is mostly idle anyway.

## C++ async vs Go goroutines

Both approaches get concurrency. They differ in how the code reads.

### C++ with Boost.Asio: explicit asynchronous callbacks

The reference C++ scanner calls something like `async_connect(socket, address, callback)`:

1. Start the connection and return **immediately**.
2. A single **event loop** (`io_context`) watches all pending connections.
3. When one finishes, the loop runs its **callback**, which records the result and starts the next port.

It's very efficient, but one simple idea ("try this port, then record the result") is
split across several callbacks, plus timers, flags to stop double handling, and a "strand"
to keep shared counters safe. That makes it hard to read and hard to get right.

### Go: write blocking code, run it in many goroutines

In Go you write plain **synchronous** code that reads top to bottom:

```go
conn, err := net.DialTimeout("tcp", addr, timeout) // simply waits for the result
```

You get concurrency by running that code in many **goroutines**:

- A goroutine is a very lightweight thread managed by the Go runtime. It starts at a few KB of memory, so 1,000+ of them is fine.
- You start one with the `go` keyword: `go worker(...)`.
- When a goroutine waits on the network, the runtime **parks** it and runs another one.
- Under the hood the Go runtime uses the OS's **asynchronous** network I/O (the "netpoller") for you.

So the Go scanner is **concurrent** and **asynchronous under the hood**, while every
function still reads like normal step-by-step code.

### What this means for our stages

| Stage | What you write |
|-------|----------------|
| 1 | `ScanPort`: plain blocking code for **one** port |
| 4 | the **same** `ScanPort`, now called from many goroutines (a worker pool) connected by **channels** |

You never rewrite the scanning logic to make it concurrent. You only change *how many
copies of it run at once*. That separation is the main advantage of Go for this kind of tool.
