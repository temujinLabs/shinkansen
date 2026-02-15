# Shinkansen — Enhancement Plan (MVP)

## Problem Statement

Dev teams stuck on Jira hate the slow, bloated UI but can't switch to Linear due to enterprise lock-in (compliance, integrations, admin policies). They want Linear's keyboard-driven speed while keeping Jira as the source of truth. The Jira web UI adds friction to every interaction — page loads, clicks, dropdowns, modal waits — and this friction compounds across hundreds of daily interactions.

## Competitive Landscape

**Existing Jira CLI tools:**
- **jira-cli (ankitpokhrel)** — Feature-rich Go CLI. ~4K GitHub stars. Interactive issue creation, sprint management, board views. The best existing option. Limitation: still requires multiple commands for common workflows, no TUI, setup is complex.
- **go-jira** — Simpler Go CLI. Mature but less maintained. Template-based output. Good for scripting, not for interactive use.
- **Atlassian ACLI** — Official Atlassian CLI. Admin/automation focused, not developer-workflow focused. Clunky, enterprise-oriented.
- **Appfire Jira CLI** — Marketplace product. Paid. Focused on bulk operations and migrations, not daily developer use.

**Why Linear feels fast:**
- Built from scratch with modern frontend (no legacy). Loads in ~half the time of Jira.
- Keyboard-driven — Cmd+K does everything. Minimal mouse usage needed.
- Opinionated workflows — less configuration means fewer UI elements, faster rendering.
- Optimistic updates — UI reflects changes before server confirms.
- Offline-capable architecture — local-first sync.

**Popular TUI tools developers already love:**
- **lazygit** — Git TUI. 50K+ stars. Proves developers want terminal UIs for daily tools.
- **k9s** — Kubernetes TUI. 25K+ stars. Complex data, simple interface.
- **gh** — GitHub CLI. Widely adopted. Proves CLI for issue tracking works.
- **tig** — Git log viewer TUI. Simple, fast, beloved.

**Key insight:** No one has built the "lazygit for Jira" — a fast, keyboard-driven TUI that makes Jira feel instant.

## Shinkansen's Differentiation

1. **TUI-first, not just CLI** — Interactive terminal UI with panels, not just commands. Think lazygit but for Jira.
2. **Sub-second everything** — Aggressive local caching. Common operations feel instant.
3. **Zero-config start** — `shinkansen login` → immediately usable. No YAML config files.
4. **Opinionated shortcuts** — The 10 things devs do most in Jira, optimized to 1-2 keystrokes each.
5. **Works alongside Jira UI** — Not a replacement. A fast lane for developers.

## Core MVP Features (5 features, no more)

### 1. Landing Page + Email Capture
- Headline: "Jira at the speed of thought"
- Speed comparison demo (animated GIF: Jira UI vs Shinkansen for same task)
- Email capture for early access / beta download
- "Works with your existing Jira Cloud instance"

### 2. Authentication + Quick Setup
- `shinkansen login` — OAuth or API token flow
- Auto-detect Jira instance, projects, boards
- Store credentials securely in OS keychain
- Ready to use in <30 seconds

### 3. Interactive TUI Dashboard
- **My Issues** panel — assigned to me, sorted by priority
- **Sprint Board** panel — current sprint, kanban columns
- **Quick Actions** — single keypress operations:
  - `Enter` — Open issue detail
  - `m` — Move issue (status transition)
  - `a` — Assign to someone
  - `c` — Add comment
  - `n` — Create new issue
  - `/` — Search (fuzzy find across all issues)
  - `q` — Quit
- Vim-style navigation (j/k up/down, h/l panels)

### 4. Fast Issue Operations
- **Create**: `n` → type title → select type/priority → done (3 keystrokes + typing)
- **Transition**: `m` → select status → done (2 keystrokes)
- **Comment**: `c` → type → Enter (2 keystrokes + typing)
- **Search**: `/` → fuzzy search → Enter to jump (instant, searches local cache)
- **View**: Full issue detail in a scrollable pane (description, comments, attachments, links)

### 5. Local Cache + Offline View
- Sync issues on startup and periodically
- Search works against local cache (instant results)
- View issues offline
- Queue changes when offline, sync when back online
- Cache invalidation on Jira webhook or manual refresh (`r`)

## Technical Architecture

### Language Choice: Go or Rust

**Go (recommended for MVP):**
- jira-cli already exists in Go (can study patterns)
- Excellent TUI libraries: Bubble Tea (Charm), tview
- Fast compilation, single binary distribution
- Good HTTP client ecosystem for Jira API
- Cross-platform without hassle

**Rust (if performance is critical later):**
- Ratatui (TUI library) is excellent
- Faster runtime, but slower dev velocity for MVP

**Recommendation: Go + Bubble Tea**

### Jira API Strategy

```
Jira Cloud REST API v3
├── Authentication: OAuth 2.0 (3LO) or API Token + Basic Auth
├── Key Endpoints:
│   ├── GET /rest/api/3/search (JQL search)
│   ├── GET /rest/api/3/issue/{key} (issue detail)
│   ├── POST /rest/api/3/issue (create)
│   ├── PUT /rest/api/3/issue/{key} (update)
│   ├── POST /rest/api/3/issue/{key}/transitions (move)
│   ├── POST /rest/api/3/issue/{key}/comment (comment)
│   ├── GET /rest/agile/1.0/board/{id}/sprint (sprints)
│   └── GET /rest/agile/1.0/sprint/{id}/issue (sprint issues)
├── Rate Limits: ~10 req/sec for Cloud
├── Pagination: startAt + maxResults (default 50)
└── Webhooks: Available for real-time sync
```

### Caching Strategy

```
Local SQLite cache:
├── issues (key, summary, status, assignee, priority, updated_at, raw_json)
├── projects (id, key, name)
├── sprints (id, name, state, board_id)
├── transitions (issue_key, transition_id, name)
└── sync_log (last_sync, items_synced, duration_ms)

Sync approach:
1. On startup: Fetch issues updated since last_sync (JQL: updated >= "last_sync_time")
2. Background: Poll every 60 seconds for changes
3. On demand: User presses 'r' to force refresh
4. Cache-first: All reads hit cache, API only for writes and sync
```

### Application Architecture

```
shinkansen/
├── cmd/
│   └── shinkansen/main.go       # Entry point
├── internal/
│   ├── tui/                     # Bubble Tea TUI components
│   │   ├── app.go               # Main app model
│   │   ├── board.go             # Sprint board view
│   │   ├── issues.go            # Issue list view
│   │   ├── detail.go            # Issue detail view
│   │   ├── search.go            # Fuzzy search
│   │   └── styles.go            # Lip Gloss styles
│   ├── jira/                    # Jira API client
│   │   ├── client.go            # HTTP client with auth
│   │   ├── issues.go            # Issue operations
│   │   ├── sprints.go           # Sprint operations
│   │   └── search.go            # JQL search
│   ├── cache/                   # Local SQLite cache
│   │   ├── store.go             # Cache read/write
│   │   └── sync.go              # Sync logic
│   └── config/                  # Config + keychain
│       ├── config.go            # Settings
│       └── auth.go              # Credential storage
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

### Distribution

- **Homebrew**: `brew install shinkansen`
- **Go install**: `go install github.com/temujinlabs/shinkansen@latest`
- **GitHub Releases**: Pre-built binaries for macOS (arm64/amd64), Linux, Windows
- **No Docker needed** — single binary

## Landing Page (Web Component)

The landing page and email capture use the standard Lab template (FastAPI + Tailwind). The CLI tool itself is a separate Go binary.

```
Web (FastAPI):
├── Landing page (/)
├── Download page (/download)
├── Docs (/docs)
└── API for email capture

CLI (Go binary):
├── Distributed via Homebrew/GitHub Releases
├── Connects directly to user's Jira Cloud
└── No dependency on our servers (except telemetry opt-in)
```

## Validation Metrics

### Core Feature to Validate
Will developers actually use a TUI for Jira instead of the web UI?

### Week 1 Targets
- 50+ landing page visitors
- 15+ email signups (dev tools convert higher)
- 5+ CLI downloads
- 100+ total commands executed across all users

### Week 4 Go/No-Go
- 200+ visitors
- 30+ downloads
- 40%+ daily active users (DAU/downloads — people using it as part of their workflow)
- 500+ commands executed
- 1+ "I can't go back to Jira UI" feedback

### Kill Criteria
- If <40% of downloaders use CLI daily after week 2 → value prop isn't strong enough
- If users keep opening Jira web for basic operations → TUI isn't saving time

## Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Jira API rate limits (10 req/sec) | Slow sync | Aggressive caching, batch requests, only fetch deltas |
| Jira Server vs Cloud differences | Fragmented user base | MVP targets Cloud only. Server support later. |
| OAuth setup is complex for users | Low activation | Support API token (simpler) as primary auth, OAuth as option |
| Enterprise Jira has custom fields/workflows | TUI can't handle all cases | Focus on standard fields. Allow raw JSON view for custom fields. |
| Go + Bubble Tea learning curve | Slower MVP development | Well-documented library, active community, many examples |
| Existing jira-cli (ankitpokhrel) is good enough | Users don't switch | Differentiate on TUI experience, not just CLI commands. lazygit-style, not gh-style. |

## Cost Estimate (Monthly)

| Item | Cost |
|------|------|
| VPS for landing page | $6 (shared) |
| Domain | $0 (shared) |
| GitHub Actions CI/CD | $0 (free tier) |
| Homebrew tap hosting | $0 (GitHub) |
| **Total** | **~$6/mo** |

The CLI connects directly to the user's Jira — no server costs for the core product.

## Pricing Model

| Plan | Price | Includes |
|------|-------|---------|
| Free (Open Source) | $0 | Full TUI, single Jira instance, local cache |
| Pro | $9/mo | Multi-instance, team sync, priority support |
| Team | $29/mo | Admin dashboard, usage analytics, SSO |

**Note:** Dev tools need to be free/cheap to gain adoption. The free tier IS the product. Revenue comes from team features later.
