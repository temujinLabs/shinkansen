# Shinkansen

Sub-second CLI/TUI for Jira — bullet train speed + precision.

## Quick Start

```bash
cd shinkansen/

# Build locally (requires Go 1.22+)
make build
./shinkansen login
./shinkansen

# Build via Docker (no Go needed)
make docker-extract
./shinkansen login
./shinkansen
```

## Stack

- **Language**: Go 1.22
- **TUI**: Bubble Tea + Lip Gloss + Bubbles (charmbracelet)
- **Cache**: SQLite via modernc.org/sqlite (pure Go, no CGO)
- **API**: Jira Cloud REST API v3

## Key Commands

```
shinkansen login     — Configure Jira credentials
shinkansen           — Launch TUI
shinkansen --version — Print version
```

## TUI Keybindings

```
j/k     Navigate up/down
h/l     Switch panels (issues | board)
Enter   Open issue detail
m       Move issue (status transition)
c       Add comment
n       Create new issue
/       Fuzzy search
r       Refresh/sync
?       Help
q       Quit
```

## Architecture

```
cmd/shinkansen/main.go     Entry point
internal/tui/              Bubble Tea views
internal/jira/             Jira API client
internal/cache/            SQLite cache layer
internal/config/           Config + auth
```

## Conventions

- Cache-first reads (sub-100ms)
- Delta sync (only fetch updated issues)
- Config at ~/.config/shinkansen/
- Single binary, no dependencies
