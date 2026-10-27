# Stage 1: Scan ONE port

Goal: a function that takes a host and a port and says **open**, **closed** or **filtered**.

Everything later (lists of ports, concurrency, banners) is this one function called many times.
Get it right and the rest is plumbing.

---

## Step 1: Analyse the problem before typing

Answer these on paper first. The answers *are* the design.

| Question | Answer |
|----------|--------|
| What goes **in**? | a host (`"127.0.0.1"` or `"scanme.nmap.org"`), a port (`22`), how long to wait (`1s`) |
| What comes **out**? | one of three words: `open`, `closed`, `filtered` |
| How do I "knock" on a port in Go? | try to open a TCP connection: the `net` package |
| How do I tell the 3 outcomes apart? | look at the **error** the connection attempt returns (table below) |
| What can go wrong? | waiting forever → always use a timeout; leaking connections → always close them |

Mapping network events to Go results:

| On the wire | What `net.DialTimeout` gives you | State |
|-------------|----------------------------------|-------|
| SYN → SYN-ACK | `conn` and `err == nil` | **open** |
| SYN → RST | an error, quickly ("connection refused") | **closed** |
| SYN → nothing… | an error whose `Timeout()` is `true` | **filtered** |

So the whole function comes down to: **dial, then sort the error into one of three buckets.**

---

## Step 2: The Go tools you need

Read these doc pages. They're short, and reading Go docs is a skill in itself:

- [`net.DialTimeout`](https://pkg.go.dev/net#DialTimeout): `DialTimeout(network, address string, timeout time.Duration) (Conn, error)`
  - `network` is `"tcp"`
  - `address` is `"host:port"`
- [`net.JoinHostPort`](https://pkg.go.dev/net#JoinHostPort): builds `"host:port"` correctly.
  *Why not `host + ":" + port`?* IPv6 addresses contain colons (`::1`) and must be written `[::1]:80`. `JoinHostPort` handles that for you.
- [`strconv.Itoa`](https://pkg.go.dev/strconv#Itoa): `JoinHostPort` wants strings, and your port is an `int`.
- [`time.Duration`](https://pkg.go.dev/time#Duration): write `1 * time.Second` or `500 * time.Millisecond`, never a bare `1`.
- **`defer conn.Close()`**: when the port is open you now hold a real connection. Close it, or a 65,535-port scan runs out of sockets.

### The one tricky idiom: "was this error a timeout?"

Network errors in Go implement the `net.Error` interface, which has a `Timeout() bool` method.
The error you get back is *wrapped* (an error inside an error), so you unwrap it with `errors.As`:

```go
var netErr net.Error
if errors.As(err, &netErr) && netErr.Timeout() {
    // no reply in time → filtered
}
```

Read it as: *"if somewhere inside `err` there is a `net.Error`, put it in `netErr`; then ask it if it was a timeout."*

Any other error (refused, reset, unreachable...) we'll call **closed** for now. That matches
what simple scanners do: *a fast failure means something answered "no".*

---

## Step 3: Your task: write it

**File:** `internal/scanner/scanner.go`

```go
package scanner

// ScanPort tries a TCP connection to host:port and reports
// "open", "closed" or "filtered".
func ScanPort(host string, port int, timeout time.Duration) string {
    // your code
}
```

Notes:
- The name starts with a **capital S**. In Go, capitalised = **exported** = usable from other packages (your `main`). `scanPort` would be invisible to `main`.
- Instead of typing `"open"` etc. by hand in several places, define constants once:
  `const ( Open = "open"; Closed = "closed"; Filtered = "filtered" )`. A typo then becomes a compile error, not a silent bug.
- Shape of the body: build the address → dial → if no error: close and return open → if timeout: filtered → else closed.

Then update **`cmd/portscanner/main.go`** to import your package and call it for a few ports:

```go
import "github.com/<your-github-username>/tcp-port-scanner/internal/scanner"
```

Print something like `127.0.0.1:135  open`.

---

## Step 4: Test it against all three states

You need one known example of each state:

| State | Target | How |
|-------|--------|-----|
| open | `127.0.0.1` port **9000** | start a listener in a second PowerShell window (below) |
| open | `scanme.nmap.org` port **22** or **80** | Nmap's public test host |
| closed | `127.0.0.1` port **9999** | nothing listens there |
| filtered | `10.255.255.1` port **80** | a private address that nobody answers, so the SYN goes nowhere |

Temporary listener on port 9000 (stop it with `$l.Stop()`):

```powershell
$l = [System.Net.Sockets.TcpListener]::new([System.Net.IPAddress]::Loopback, 9000); $l.Start()
```

### 🔍 Experiment: Windows has a surprise for you

Time a **closed** port on a **remote** host (e.g. `scanme.nmap.org` port 23) on Windows.
It takes **~3 seconds** instead of one round trip, and with a 1 s timeout it shows up as
**filtered**. Windows retries a refused connection before reporting it. Linux and macOS don't.
Closed ports on `127.0.0.1` are **not** affected: measured at about 1 ms each.

Measure with a built binary (`go build`), not `go run`. `go run` includes compile time,
which can easily add 1–2 s and mislead you.

Try a 1 s timeout, then a 3 s timeout, and compare. This is a good thing to mention in
your README: it shows you tested and understood your tool's limits.

---

## ✅ Done when

- `go run ./cmd/portscanner` prints the right state for all three test targets
- You can explain, without notes, why a timeout means *filtered* and a fast error means *closed*
- `go vet ./...` reports nothing
- Commit: `git commit -m "Stage 1: scan a single port"`

**Then send me your `scanner.go` and `main.go` for review before Stage 2.**
