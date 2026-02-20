# Changelog

## v0.1.0 — 2026-02-20

First public release.

### Features

- **Board view** — Navigate Jira boards with columns mapped to statuses
- **Issue list** — Browse and scroll through project issues
- **Keyboard shortcuts** — Full keyboard-driven navigation (j/k, h/l, Enter, and more)
- **Cache-first reads** — All reads served from local SQLite for sub-100ms response times
- **Delta sync** — Background incremental sync fetches only issues modified since the last update
- **Transitions** — Move issues between statuses with a single keystroke (`m`)
- **Comments** — Add comments to issues inline (`c`)
- **Time logging** — Log work directly from the TUI (`t`)
- **Assign** — Assign issues to yourself (`a`)
- **Bulk operations** — Select multiple issues and apply transitions in bulk (`Space m`)
- **Search** — Fuzzy search across cached issues (`/`)
- **Open in browser** — Jump to any issue in your default browser (`o`)
- **OAuth 2.0 support** — Authenticate via OAuth 2.0 (3LO) as an alternative to API tokens
- **Single binary** — No runtime dependencies; one binary, zero setup
