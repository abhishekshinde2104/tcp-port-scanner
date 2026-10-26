# Stage 0: Setup

Goal: a Go module that builds and runs, inside a git repo that's ready to push.

## 1. What is a Go module?

A **module** is a Go project. The file `go.mod` at the root gives the project a
name (its **module path**) and records the Go version and dependencies.

The module path is usually where the code will live on GitHub:

```
github.com/<your-github-username>/tcp-port-scanner
```

This matters because your own packages are imported **by that path**. Later
`main.go` will contain:

```go
import "github.com/<your-github-username>/tcp-port-scanner/internal/scanner"
```

Pick the path now so you never have to rename imports later.

## 2. Why this folder layout?

| Folder | Why |
|--------|-----|
| `cmd/portscanner/` | Go convention: each runnable program gets a folder under `cmd/`. Its `main.go` stays *thin*: read flags, call packages, print. |
| `internal/...` | Packages under `internal/` can only be imported by this module. That signals "not a public library" and keeps you free to change them. |
| one package per job | `ports` parses, `scanner` scans, `output` prints. Each piece is small, testable and easy to explain in an interview. |

**Thinking habit:** before coding, ask *"what is this program's single job, and what are the separate sub-jobs?"* Each sub-job becomes a package or function.

## 3. Your tasks

Run these from `D:\AI\tcp_port_scanner`:

```bash
go mod init github.com/<your-github-username>/tcp-port-scanner
```

```bash
git init
```

```bash
git status
```

Now create `cmd/portscanner/main.go` yourself. It needs:

1. `package main`: the package that produces a runnable program
2. an import of `fmt`
3. `func main()` that prints `tcp-port-scanner v0.0.1`

Run it:

```bash
go run ./cmd/portscanner
```

You can delete `cmd/portscanner/.gitkeep` once `main.go` exists. `.gitkeep` files only
exist so git tracks empty folders.

## 4. First commit (you do this, under your own name)

```bash
git add .
```

```bash
git commit -m "Stage 0: project skeleton"
```

Check with `git log`: the author should be you.

## ✅ Done when

- `go run ./cmd/portscanner` prints the version line
- `git log` shows your commit and `.claude/` is not tracked
