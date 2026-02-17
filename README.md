# Shinkansen

A keyboard-driven TUI for Jira. Fast as a bullet train.

## Why

Jira is a standard — and a pain. Slow loading, click-heavy UI, feature overload. Shinkansen gives you a terminal interface that's as fast as Linear but works with your existing Jira setup.

## Features

| Feature | Key | Description |
|---------|-----|-------------|
| **Issue List** | `↑/↓` | Navigate your assigned issues |
| **Board View** | `←/→` | Kanban board (To Do / In Progress / Done) |
| **Open Detail** | `Enter` | Full issue view with description + comments |
| **Open in Browser** | `o` | Jump to the issue in Jira web |
| **Assign to Self** | `a` | One-key self-assignment |
| **Move Status** | `m` | Pick a status transition |
| **Bulk Move** | `Space` then `m` | Select multiple issues, move all at once |
| **Add Comment** | `c` | Inline comment (ADF format) |
| **Log Time** | `t` | Log work (e.g. "2h", "30m") |
| **Create Issue** | `n` | Quick new task creation |
| **Search** | `/` | Fuzzy search across cached issues |
| **Refresh** | `r` | Force sync from Jira |
| **Help** | `?` | Keyboard shortcuts reference |
| **Quit** | `q` | Exit |

## Install

### Download binary (macOS/Linux)

```bash
# macOS ARM64 (Apple Silicon)
curl -L https://github.com/temujinlabs/shinkansen/releases/latest/download/shinkansen-darwin-arm64 -o shinkansen
chmod +x shinkansen
./shinkansen login
```

### From source

```bash
go install github.com/temujinlabs/shinkansen/cmd/shinkansen@latest
```

### Docker

```bash
docker build -t shinkansen .
```

## Setup

```bash
$ ./shinkansen login
Jira URL (e.g. https://yourorg.atlassian.net): https://myorg.atlassian.net
Email: you@company.com
API Token: [paste from https://id.atlassian.com/manage-profile/security/api-tokens]

Authenticated as: Your Name (you@company.com)
Found 3 projects
Configuration saved. Run 'shinkansen' to start.
```

Then just run:

```bash
$ ./shinkansen
```

## How It Works

- **Cache-first**: All reads from local SQLite (~/.config/shinkansen/cache.db). Sub-100ms.
- **Delta sync**: Only fetches issues changed since last sync. Every 60 seconds by default.
- **Offline capable**: Browse cached issues without network.
- **Writes go direct**: Comments, transitions, assignments hit the Jira API immediately.

## Architecture

```
TUI (Bubble Tea) → Jira Client (REST API v3) → Jira Cloud
                 → SQLite Cache (local, WAL mode)
```

Single binary. No CGO. No runtime dependencies.

## Configuration

Stored at `~/.config/shinkansen/config.json`:

```json
{
  "jira_url": "https://yourorg.atlassian.net",
  "email": "you@company.com",
  "api_token": "...",
  "account_id": "...",
  "default_project": "SCRUM",
  "sync_interval": 60
}
```

## Name

Japanese bullet train — legendary speed + precision for Jira workflows.

## Credits

Built by [Guillem Rovira](https://github.com/guillemrh) at [Temujin Labs](https://temujinlabs.com).
