# Swifty

An AI coding agent for your terminal, built in Go on top of [Bubble Tea](https://github.com/charmbracelet/bubbletea).

## Requirements

- Go >= 1.26
- Node.js >= 18 (only required for the `build.mjs` release script)

## Development

Run the CLI directly from source:

```sh
go run ./cmd/swifty
```

Run the test suite:

```sh
go test ./...
```

## Remote mode

Remote mode starts a web UI server (instead of the terminal TUI) that bridges
the agent to browser clients over WebSocket. The default listen address is
`:18888`; an optional address argument overrides it.

Start from source (dev):

```sh
# default address :18888
go run ./cmd/swifty --remote

# custom address
go run ./cmd/swifty --remote :9000
```

Start from a packaged binary:

```sh
go build -o build/swifty ./cmd/swifty

# default address :18888
./build/swifty --remote

# custom address
./build/swifty --remote :9000
```

Then open the printed URL (e.g. `http://localhost:18888`) in a browser.

## Building

### Build for the current platform

```sh
go build -o build/swifty ./cmd/swifty
```

### Cross-platform compilation

Go supports cross-compilation out of the box via the `GOOS` and `GOARCH`
environment variables. `CGO_ENABLED=0` guarantees a fully static binary, and
`-ldflags "-s -w"` strips the symbol table and DWARF debug info to reduce the
binary size.

Note: Go names the x86_64 architecture `amd64` (`GOARCH=amd64`), while the
release artifacts use the `x64` label (Node.js-style naming).

```sh
# macOS (Apple Silicon)
CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/swifty-darwin-arm64 ./cmd/swifty

# macOS (Intel)
CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/swifty-darwin-x64 ./cmd/swifty

# Linux (x86_64)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/swifty-linux-x64 ./cmd/swifty

# Linux (ARM64)
CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/swifty-linux-arm64 ./cmd/swifty

# Windows (x86_64)
CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags "-s -w" -o build/swifty-windows-x64.exe ./cmd/swifty

# Windows (ARM64)
CGO_ENABLED=0 GOOS=windows GOARCH=arm64 go build -trimpath -ldflags "-s -w" -o build/swifty-windows-arm64.exe ./cmd/swifty
```

### Build all platforms at once

The repository ships a release script that compiles binaries for
`darwin`, `linux`, and `windows` (both `x64` and `arm64`) in one pass and
writes them to the `./build` directory:

```sh
node build.mjs
```

Example output:

```
build/
├── swifty-darwin-arm64
├── swifty-darwin-x64
├── swifty-linux-arm64
├── swifty-linux-x64
├── swifty-windows-arm64.exe
└── swifty-windows-x64.exe
```
