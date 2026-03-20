# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

logid (byte-logid) is an internal CLI tool for querying distributed tracing logs by Log ID (Trace ID). It supports multiple regions (us/i18n/eu/cn), outputs structured JSON, and delegates authentication to an external `byte-auth` tool.

**Internal use only - do not share externally.**

## Common Commands

```bash
make build          # Build binary with version info via ldflags
make test           # Run all tests: go test ./... -v
make build-all      # Cross-compile for darwin/linux (amd64/arm64)
make install        # Build and install to ~/.local/bin/

# Run a single test
go test ./internal/config/ -run TestParseRegion -v

# Run the tool locally
go run ./cmd/logid <trace-id> -r us
```

## Architecture

CLI built with **cobra** (`github.com/spf13/cobra`). Entry point is `cmd/logid/main.go`, which injects version info via ldflags and calls into `cmd/` package.

### Data Flow

```
User runs: logid <LOGID> -r <region>
  -> cmd/root.go:runQuery()         # orchestrates the full pipeline
    -> config.ParseRegion()         # validate region enum
    -> auth.GetToken()              # get JWT via --token flag or byte-auth CLI
    -> config.NewAppConfig()        # ensure ~/.config/logid/ exists
    -> config.Load()                # load filters.json (sanitization regex rules)
    -> filter.NewMessageSanitizer() # compile regex patterns
    -> filter.NewKeywordFilter()    # setup keyword matching (case-insensitive, OR)
    -> query.NewClient().Query()    # HTTP call to log service, then sanitize+filter+truncate
    -> JSON output to stdout
```

### Key Packages

- **cmd/** - Cobra command definitions. `root.go` contains the main query logic; `config.go` manages filter rules; `update.go` handles self-update; `version.go` prints version.
- **internal/auth/** - Authentication provider. Shells out to `byte-auth token --region <r> --raw` or uses manual `--token`.
- **internal/config/** - Region enum/config mapping, app config (`~/.config/logid/`), filter rules (`filters.json`).
- **internal/filter/** - Two-stage local filtering: `sanitizer.go` removes noise via regex, `keyword.go` filters entries by keywords.
- **internal/query/** - HTTP client for the log service API. `types.go` defines request/response structs. `client.go` handles query, sanitization, keyword filtering, and message truncation.
- **internal/updater/** - Self-update via `go install`.

### Three-Layer Filtering + Truncation

1. **PSM filter** (server-side, `--psm`) - only return logs from specified services
2. **Message sanitization** (local, `filters.json`) - regex-based removal of noise from `_msg`
3. **Keyword filter** (local, `--keyword`) - keep only entries matching keywords (before truncation)
4. **Message truncation** (local, `--max-len`, default 1000) - truncate long messages

### Config Files

Runtime config lives in `~/.config/logid/filters.json` (auto-created on first run).

## Language

All user-facing strings, error messages, and documentation are in Chinese (Simplified).
