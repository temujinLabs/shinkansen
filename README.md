# Shinkansen

Keyboard-driven TUI for Jira. Fast as a bullet train.

## Features

- **Issue list view** with inline status and priority
- **Kanban board** with columns per status
- **Fuzzy search** across all issues
- **JQL filter** with saved filter history
- **Issue creation** without leaving the terminal
- **SQLite cache** for sub-100ms reads
- **Delta sync** fetches only updated issues

## Keybindings

| Key | Action |
|-----|--------|
| `j` / `k` | Navigate up / down |
| `h` / `l` | Switch panels (issues, board) |
| `Enter` | Open issue detail |
| `m` | Move issue (status transition) |
| `c` | Add comment |
| `n` | Create new issue |
| `f` | JQL filter |
| `p` | Switch project |
| `/` | Fuzzy search |
| `r` | Refresh / sync |
| `?` | Help |
| `q` | Quit |

## Installation

### Homebrew

```bash
brew tap temujinlabs/tap
brew install shinkansen
```

### Manual download

Download the latest binary for your platform from
[GitHub Releases](https://github.com/temujinLabs/shinkansen/releases).

Available builds:

- `shinkansen-darwin-amd64` (macOS Intel)
- `shinkansen-darwin-arm64` (macOS Apple Silicon)
- `shinkansen-linux-amd64` (Linux x86_64)
- `shinkansen-linux-arm64` (Linux ARM64)

Make it executable and move it to your PATH:

```bash
chmod +x shinkansen-*
mv shinkansen-* /usr/local/bin/shinkansen
```

## Requirements

- Jira Cloud account with an API token
- Generate a token at [id.atlassian.com/manage-profile/security/api-tokens](https://id.atlassian.com/manage-profile/security/api-tokens)

## Configuration

On first run, Shinkansen will prompt for your Jira credentials:

- Jira base URL (e.g. `https://yourcompany.atlassian.net`)
- Email address
- API token

Credentials are stored locally in `~/.config/shinkansen/config.json`.

## Links

- Homepage: [shinkansen.temujinlabs.com](https://shinkansen.temujinlabs.com)

## License

MIT. See [LICENSE](LICENSE) for details.

---

Built by [Temujin Labs](https://temujinlabs.com)
