# Agent Deck

Terminal session manager for AI coding agents. Built with Go + Bubble Tea.

**Version**: 0.3.0

---

## CRITICAL: Data Protection Rules

**THIS SECTION MUST NEVER BE DELETED OR IGNORED**

### tmux Session Loss Prevention

**Root Cause (2025-12-09 Incident)**: All 37 sessions were lost when the tmux server restarted. The session metadata in `sessions.json` remained intact, but the underlying tmux sessions were destroyed.

**NEVER DO THESE THINGS:**
1. **NEVER run `tmux kill-server`** - This destroys ALL agent-deck sessions instantly
2. **NEVER run `tmux kill-session` with patterns** - Commands like `tmux ls | grep agentdeck | xargs tmux kill-session` will DESTROY ALL SESSIONS
3. **NEVER quit Terminal.app/iTerm completely** while sessions are running - the tmux server may not survive
4. **NEVER restart macOS** without exporting important session outputs first
5. **NEVER run tests that might interfere with production tmux sessions**
6. **NEVER run cleanup commands** that target "agentdeck" or session patterns - Claude in dangerous mode may do this autonomously

**⚠️ 2025-12-10 Incident (Root Cause Confirmed):**
A Claude Code instance running in **dangerous mode** (`bypass permissions on`) autonomously ran:
```bash
tmux ls | grep agentdeck | cut -d: -f1 | xargs -I{} tmux kill-session -t {}
```
This killed ALL 40 sessions instantly. The session was trying to "clean up demo sessions" but matched ALL agent-deck sessions.

**Recovery Procedures:**
```bash
# Session logs are preserved in ~/.agent-deck/logs/
# Each session has: agentdeck_{name}_{id}.log

# To see what was in a session:
tail -500 ~/.agent-deck/logs/agentdeck_<session-name>_<id>.log

# Session metadata backups (rolling 3 generations):
~/.agent-deck/profiles/default/sessions.json.bak
~/.agent-deck/profiles/default/sessions.json.bak.1
~/.agent-deck/profiles/default/sessions.json.bak.2
```

**Session Log Files Location:**
- All tmux output is logged to `~/.agent-deck/logs/`
- Logs persist even when tmux sessions are killed
- Use these to recover conversation history

---

## Quick Start

```bash
make build      # Build to ./build/agent-deck (updates binary instantly via symlink)
make test       # Run tests
```

### Development Setup (one-time)

A symlink is configured so `make build` instantly updates the system binary:

```bash
# /usr/local/bin/agent-deck → /Users/ashesh/claude-deck/build/agent-deck
# Already set up! Just run: make build
```

If symlink breaks, recreate with:
```bash
sudo ln -sf /Users/ashesh/claude-deck/build/agent-deck /usr/local/bin/agent-deck
```

## CLI Commands

```bash
# Interactive TUI
agent-deck

# Add session from path
agent-deck add /path/to/project
agent-deck add . -t "My Session" -g "work" -c claude

# List sessions
agent-deck list
agent-deck list --json

# Remove session
agent-deck remove <id|title>

# Quick status check (no TUI)
agent-deck status              # Compact: "2 waiting • 5 running • 3 idle"
agent-deck status -v           # Verbose: detailed list by status
agent-deck status -q           # Quiet: just waiting count (for scripts/prompts)
agent-deck status --json       # JSON: machine-readable output

# Version/Help
agent-deck version
agent-deck help
```

### Add Command Flags
| Flag | Description |
|------|-------------|
| `-t`, `--title` | Custom session title (defaults to folder name) |
| `-g`, `--group` | Group path (defaults to parent folder) |
| `-c`, `--cmd` | Command: claude, gemini, aider, codex, cursor, or custom |

## Project Structure

```
cmd/agent-deck/main.go        # Entry point + CLI subcommands
internal/
├── ui/                       # TUI (Bubble Tea)
│   ├── home.go               # Main model (1375 lines), Update(), View()
│   ├── styles.go             # Tokyo Night colors, lipgloss styles
│   ├── group_dialog.go       # Group create/rename/move dialog
│   ├── newdialog.go          # New session dialog
│   ├── help.go               # Help overlay (keyboard shortcuts)
│   ├── search.go             # Search component with status filters
│   ├── menu.go               # Menu rendering
│   ├── preview.go            # Preview pane
│   ├── tree.go               # Tree rendering
│   └── list.go               # List navigation
├── session/                  # Data layer
│   ├── instance.go           # Session struct, Status enum
│   ├── groups.go             # GroupTree, nested hierarchy
│   ├── storage.go            # JSON persistence
│   ├── config.go             # Profile management (JSON)
│   ├── userconfig.go         # User config (TOML) - custom tools
│   └── discovery.go          # Import existing tmux sessions
└── tmux/                     # tmux integration
    ├── tmux.go               # Session management, status detection (1000+ lines)
    ├── detector.go           # Tool/prompt detection (362 lines)
    └── pty.go                # PTY attach/detach (226 lines)
```

## Keyboard Shortcuts

### Navigation
| Key | Action |
|-----|--------|
| `j` / `↓` | Move down |
| `k` / `↑` | Move up |
| `h` / `←` | Collapse group (or parent if on session) |
| `l` / `→` / `Tab` | Toggle expand/collapse |

### Session Actions
| Key | Action |
|-----|--------|
| `Enter` | Attach to session OR toggle group |
| `n` | New session (inherits current group) |
| `S` | **Restart errored session** (recreate tmux session) |
| `G` / `R` | Rename session or group |
| `m` | Move session to different group |
| `d` | Delete session or group |
| `u` | Mark unread (idle → waiting, gray → yellow) |
| `K` / `Shift+↑` | Move item up in order |
| `J` / `Shift+↓` | Move item down in order |

### Fork Session (Claude Only)
| Key | Action |
|-----|--------|
| `f` | Quick fork - creates new session with forked Claude context |
| `F` | Fork with dialog - customize name and group |

**Requirements:**
- Claude Code must be running in the session
- Session must have a valid `lastSessionId` in Claude's config

**Configuration:**
If using a custom Claude profile (e.g., dual account setup), set in `~/.agent-deck/config.toml`:
```toml
[claude]
config_dir = "~/.claude-work"
```

### Group Actions
| Key | Action |
|-----|--------|
| `g` | New group (subgroup if cursor on group) |
| `Tab` / `l` | Toggle expand/collapse |
| `h` | Collapse group |

### View Actions
| Key | Action |
|-----|--------|
| `/` | Search sessions (fuzzy match) |
| `?` | Show keyboard shortcuts help |
| `i` | Import existing tmux sessions |
| `r` | Refresh sessions |

### Search Filters
| Query | Action |
|-------|--------|
| `waiting` | Show only waiting sessions |
| `running` | Show only running sessions |
| `idle` | Show only idle sessions |
| `<text>` | Fuzzy search title/path/tool |

### Global
| Key | Action |
|-----|--------|
| `Ctrl+Q` | Detach from attached session (tmux keeps running) |
| `q` / `Ctrl+C` | Quit application |

## Core Concepts

### Sessions (internal/session/instance.go)
```go
type Instance struct {
    ID          string    // Unique hex-8 + timestamp
    Title       string    // Display name
    ProjectPath string    // Working directory
    GroupPath   string    // Hierarchy path: "parent/child"
    Command     string    // Command to run
    Tool        string    // "claude", "gemini", "aider", "codex", "shell"
    Status      Status    // running, waiting, idle, error
    CreatedAt   time.Time
    tmuxSession *tmux.Session
}
```

### Status Indicators
| Status | Symbol | Color | Meaning |
|--------|--------|-------|---------|
| Running | `●` | Green (#9ece6a) | Busy indicator detected OR content changed within 2s |
| Waiting | `◐` | Yellow (#e0af68) | Stopped, needs attention (unacknowledged) |
| Idle | `○` | Gray (#565f89) | Stopped, user has seen it (acknowledged) |
| Error | `✕` | Red (#f7768e) | Session doesn't exist |

### Groups (internal/session/groups.go)
- **Path-based hierarchy**: `"parent/child/grandchild"`
- **`CreateGroup(name)`** / **`CreateSubgroup(parentPath, name)`**
- **`DeleteGroup(path)`** - cascades to all subgroups, moves sessions to default
- **`RenameGroup(path, newName)`** - updates all sessions + subgroups
- **`Flatten()`** - converts tree to flat list for cursor navigation
- **Empty groups persist** until explicitly deleted

### Storage (~/.agent-deck/sessions.json)
```json
{
  "instances": [
    {
      "id": "a1b2c3d4-1701234567",
      "title": "My Project",
      "project_path": "/home/user/projects/myapp",
      "group_path": "projects",
      "command": "claude",
      "tool": "claude",
      "status": "waiting",
      "created_at": "2025-12-03T12:00:00Z",
      "tmux_session": "agentdeck_my-project_a1b2c3d4"
    }
  ],
  "groups": [
    {"name": "Projects", "path": "projects", "expanded": true, "order": 0}
  ],
  "updated_at": "2025-12-03T14:30:00Z"
}
```

### User Configuration (~/.agent-deck/config.toml)

Optional TOML config for custom tools and preferences:

```toml
# Custom tool definitions
[tools.my-ai]
command = "my-ai-assistant"
icon = "🧠"
busy_patterns = ["thinking...", "processing..."]

[tools.copilot]
command = "gh copilot"
icon = "🤖"
busy_patterns = ["Generating..."]
```

| Field | Description |
|-------|-------------|
| `command` | Shell command to run the tool |
| `icon` | Emoji/symbol shown in TUI |
| `busy_patterns` | Strings that indicate tool is processing |

## UI Layout

```
┌─────────────────────────────────────────────────────────────────┐
│ [●│◐│○] Agent Deck           10 groups • 25 sessions           │
├─────────────────────────┬───────────────────────────────────────┤
│ SESSIONS                │ PREVIEW                               │
│                         │                                       │
│ ▼ Projects (5)      ●1  │ Session: My Project                  │
│   ├─ Project A      ○   │ Status: ● running                    │
│   ├─ Project B      ◐   │ 📁 /home/user/my-project             │
│   └─ Project C      ○   │ [claude] [projects]                  │
│                         │                                       │
│ ▼ Work (8)          ◐3  │ ─── Terminal Output ───              │
│   ├─ Task 1         ●   │ I'll help you with...                │
│   └─ Task 2         ○   │ > _                                  │
├─────────────────────────┴───────────────────────────────────────┤
│ [↑↓] Navigate [Enter] Attach [n] New [g] Group [d] Delete [q]  │
└─────────────────────────────────────────────────────────────────┘
```

## Dialogs

### New Session Dialog (newdialog.go)
- **Session name** - text input
- **Project path** - text input (~ expands to home)
- **Command** - selector (none, claude, gemini, aider, codex, custom)
- Auto-inherits parent group from cursor position

### Group Dialog (group_dialog.go)
4 modes:
1. **GroupDialogCreate** - create root or subgroup
2. **GroupDialogRename** - rename group (updates all paths)
3. **GroupDialogMove** - select target group for session
4. **GroupDialogRenameSession** - rename individual session

### Search (search.go)
- Fuzzy match on title, path, tool
- `j/k` or `↑/↓` to navigate
- `Enter` to select, `Esc` to cancel

## tmux Integration

### Session Naming
- Format: `agentdeck_{sanitized-title}_{unique-id}`
- Example: `agentdeck_my-project_a1b2c3d4`

### PTY Attachment (pty.go)
- **Ctrl+Q** (ASCII 17) detaches without killing session
- Window resize (SIGWINCH) forwarded to tmux
- `AcknowledgeWithSnapshot()` called on detach (baselines content hash)

## Status Detection System

### Architecture (4 Layers)

```
┌─────────────────────────────────────────────────────────────┐
│  Layer 4: UI (home.go, styles.go)                           │
│  • Tick loop (500ms) calls UpdateStatus()                   │
│  • Renders ●/◐/○/✕ symbols with colors                      │
│  • Calls AcknowledgeWithSnapshot() on detach                │
├─────────────────────────────────────────────────────────────┤
│  Layer 3: Storage (storage.go)                              │
│  • Persists Status enum to JSON                             │
│  • Restores via ReconnectSessionWithStatus()                │
│  • Calls UpdateStatus() immediately after load              │
├─────────────────────────────────────────────────────────────┤
│  Layer 2: Instance (instance.go)                            │
│  • Maps tmux status → UI status enum                        │
│  • "active"→Running, "waiting"→Waiting, "idle"→Idle         │
├─────────────────────────────────────────────────────────────┤
│  Layer 1: tmux Session (tmux.go)                            │
│  • StateTracker with hash, cooldown, acknowledged           │
│  • GetStatus() algorithm (7 steps)                          │
│  • Content normalization for stable hashing                 │
└─────────────────────────────────────────────────────────────┘
```

### Data Flow

```
STARTUP:
  storage.Load() → ReconnectSessionWithStatus(savedStatus)
       ↓
  StateTracker { hash="", acknowledged=based_on_status }
       ↓
  inst.UpdateStatus() ──immediately──→ prevents flicker

TICK LOOP (500ms):
  inst.UpdateStatus() → tmuxSess.GetStatus()
       ↓
  CapturePane() → hasBusyIndicator() → normalizeContent()
       ↓
  Compare hash, check cooldown, check acknowledged
       ↓
  Return: "active" | "waiting" | "idle" | "inactive"

USER DETACHES (Ctrl+Q):
  AcknowledgeWithSnapshot()
       ↓
  lastHash = current_hash, acknowledged = true
       ↓
  Next tick: hash same + acknowledged → "idle" (GRAY)
```

### 3-State Model
| Status | Symbol | Color | Meaning |
|--------|--------|-------|---------|
| **active** | `●` | Green | Busy indicator OR content changed within 2s |
| **waiting** | `◐` | Yellow | Stopped, needs attention (unacknowledged) |
| **idle** | `○` | Gray | Stopped, user has seen it (acknowledged) |

### StateTracker (per session)
```go
type StateTracker struct {
    lastHash       string    // SHA256 of normalized content
    lastChangeTime time.Time // When content last changed (for 2s cooldown)
    acknowledged   bool      // User has "seen" this state (yellow vs gray)
}
```

### GetStatus() Algorithm
```
1. Session doesn't exist? → "inactive"
2. Busy indicator present? → "active" (GREEN) ← catches "esc to interrupt", spinners
3. New session (nil tracker)? → init, "idle" (GRAY) ← no yellow flash
4. Restored session (empty hash)? → set hash, respect saved acknowledged state
5. Content hash changed? → "active" (GREEN)
6. Within 2s cooldown? → "active" (GREEN)
7. Cooldown expired → "idle" if acknowledged, else "waiting"
```

### Busy Indicators (hasBusyIndicator)
Detected in content to force GREEN status:
- `"esc to interrupt"` - Claude Code main indicator
- Spinner chars: `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏`
- `"Thinking..."` or `"Connecting..."` with token counts

### Content Normalization (normalizeContent)
Strips dynamic content for stable hashing to prevent false status changes:

| What | Why | Example |
|------|-----|---------|
| ANSI escape codes | Color/style changes | `\x1b[32m` (green text) |
| Control chars | Non-printing | ASCII < 32 except tab/newline |
| Braille spinners | Animated indicators | `⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏` |
| Time counters | Updates every second | `(45s · 1234 tokens)` → `(STATUS)` |
| Thinking patterns | Dynamic timing | `Thinking... (45s)` → `Thinking...` |
| Trailing whitespace | Resize artifacts | Per-line trim |
| Multiple blank lines | Cursor variations | 3+ newlines → 2 |

### Why 2-Second Cooldown?
AI agents output in bursts with micro-pauses (100-500ms). Cooldown prevents GREEN→YELLOW flickering during pauses.

### Status Persistence (storage.go)
Status survives app restarts via `ReconnectSessionWithStatus()`:

```go
// Saved status → StateTracker initialization
"idle"    → acknowledged=true,  cooldown expired → GRAY on first poll
"waiting" → acknowledged=false, cooldown expired → YELLOW on first poll
"active"  → acknowledged=false, cooldown expired → YELLOW (will turn GREEN if busy)
```

**Key**: `UpdateStatus()` is called immediately after load (not waiting for first tick) to prevent the saved status flashing before the real status is calculated.

## Tool Detection (detector.go)

### Detection Order
1. **Command string** (most reliable): claude, gemini, aider, codex
2. **Content parsing** (fallback): regex patterns for each tool
3. **Default**: "shell"

### Tool Icons
| Tool | Icon |
|------|------|
| claude | 🤖 |
| gemini | ✨ |
| aider | 🔧 |
| codex | 💻 |
| shell | 🐚 |

### 30-Second Cache
Tool detection is expensive (content parsing). Results cached for 30s. Use `ForceDetectTool()` to bypass.

## Prompt Detection (detector.go)

Detects when tool is waiting for user input.

### Claude Code
**Busy indicators (checked first)**:
- "esc to interrupt", spinner chars (⠋⠙⠹⠸⠼⠴⠦⠧⠇⠏)
- "Thinking..." or "Connecting..." with token counts

**Permission prompts**:
- "No, and tell Claude what to do differently"
- "Yes, allow once", "Allow always"
- Box dialogs: "│ Do you want", "│ Allow"
- Trust/MCP prompts

**Input prompt** (--dangerously-skip-permissions):
- Last line is just ">" or "> "

### Other Tools
- **Aider**: "(Y)es/(N)o", "aider>", or ending with ">"
- **Gemini**: "Yes, allow once", "gemini>", or ending with ">"
- **Codex**: "codex>", "Continue?", or ending with ">"
- **Shell**: "$", "#", "%", "❯", "➜", ">" prompts

## Development

### Add Keyboard Shortcut
1. `internal/ui/home.go` → `Update()` method, add `case "key":`
2. Update `renderHelpBar()` for context-aware hints

### Add Dialog
1. Create `internal/ui/mydialog.go` with `Show()`, `Hide()`, `IsVisible()`, `Update()`, `View()`
2. Add to `Home` struct, init in `NewHome()`
3. Check `IsVisible()` in `Home.Update()` and `Home.View()`

### Modify Groups
- Logic: `internal/session/groups.go`
- Dialog: `internal/ui/group_dialog.go`
- Render: `home.go` → `renderGroupItem()`, `renderSessionItem()`

## Testing

```bash
go test ./...                        # All tests
go test ./internal/session/... -v    # Session tests
go test ./internal/ui/... -v         # UI tests
go test ./internal/tmux/... -v       # tmux tests
```

## Release & Distribution

### Supported Platforms
| OS | Architecture | Notes |
|----|--------------|-------|
| macOS | amd64, arm64 | Full support |
| Linux | amd64, arm64 | Full support |
| Windows | via WSL | Use Linux binary in WSL (tmux required) |

### Release Pipeline

```
┌─────────────────────────────────────────────────────────────────┐
│                        RELEASE WORKFLOW                          │
├─────────────────────────────────────────────────────────────────┤
│                                                                  │
│  1. Update Version                                               │
│     └─ cmd/agent-deck/main.go → const Version = "X.Y.Z"         │
│                                                                  │
│  2. Commit & Tag                                                 │
│     └─ git tag vX.Y.Z && git push origin vX.Y.Z                 │
│                                                                  │
│  3. GitHub Actions Triggered (.github/workflows/release.yml)    │
│     ├─ Checkout code                                            │
│     ├─ Setup Go 1.24                                            │
│     ├─ Run tests                                                │
│     └─ Run GoReleaser                                           │
│                                                                  │
│  4. GoReleaser (.goreleaser.yaml)                               │
│     ├─ Build binaries (linux/darwin × amd64/arm64)              │
│     ├─ Create tar.gz archives                                   │
│     ├─ Generate checksums.txt                                   │
│     ├─ Create GitHub Release with assets                        │
│     └─ Update Homebrew tap formula                              │
│                                                                  │
│  5. Distribution Channels                                        │
│     ├─ GitHub Releases (direct download)                        │
│     ├─ Homebrew (asheshgoplani/tap/agent-deck)                  │
│     ├─ go install (from source)                                 │
│     └─ install.sh (curl one-liner)                              │
│                                                                  │
└─────────────────────────────────────────────────────────────────┘
```

### How to Release a New Version

```bash
# 1. Update version in code
# Edit cmd/agent-deck/main.go: const Version = "X.Y.Z"

# 2. Commit the version bump
git add cmd/agent-deck/main.go
git commit -m "chore: bump version to vX.Y.Z"
git push origin main

# 3. Create and push tag (triggers release)
git tag vX.Y.Z
git push origin vX.Y.Z

# 4. Monitor release at:
# https://github.com/asheshgoplani/agent-deck/actions
```

### CI Workflows

#### CI (.github/workflows/ci.yml)
Runs on every push/PR to main:
- **test**: Build + run all tests (ubuntu with tmux)
- **lint**: golangci-lint static analysis

#### Release (.github/workflows/release.yml)
Triggered by version tags (`v*`):
- Runs full test suite
- Executes GoReleaser for cross-compilation
- Creates GitHub Release with binaries
- Updates Homebrew formula

### Installation Methods

#### 1. Quick Install Script (Recommended)
```bash
curl -fsSL https://raw.githubusercontent.com/asheshgoplani/agent-deck/main/install.sh | bash

# With options:
curl -fsSL .../install.sh | bash -s -- --dir /usr/local/bin --version v0.2.0
```

Features:
- Auto-detects OS (darwin/linux) and arch (amd64/arm64)
- Checks for tmux dependency
- Installs to `~/.local/bin` by default
- Validates installation

#### 2. Homebrew (macOS/Linux)
```bash
brew install asheshgoplani/tap/agent-deck
```

Requires: `HOMEBREW_TAP_GITHUB_TOKEN` secret for auto-updating tap.

#### 3. Go Install (From Source)
```bash
go install github.com/asheshgoplani/agent-deck/cmd/agent-deck@latest
# or specific version:
go install github.com/asheshgoplani/agent-deck/cmd/agent-deck@v0.1.0
```

#### 4. Manual Download
Download from [GitHub Releases](https://github.com/asheshgoplani/agent-deck/releases):
```bash
# Example for macOS ARM64
curl -LO https://github.com/asheshgoplani/agent-deck/releases/download/v0.1.0/agent-deck_0.1.0_darwin_arm64.tar.gz
tar -xzf agent-deck_0.1.0_darwin_arm64.tar.gz
sudo mv agent-deck /usr/local/bin/
```

### GoReleaser Configuration

Key settings in `.goreleaser.yaml`:

| Setting | Value | Purpose |
|---------|-------|---------|
| `CGO_ENABLED=0` | Static binaries | No libc dependency |
| `ldflags -s -w` | Strip debug info | Smaller binaries |
| `ldflags -X main.Version` | Inject version | Runtime version display |
| `ignore: windows` | Skip Windows | tmux not available |
| `brews` | Homebrew formula | Auto-update tap repo |

### Build Artifacts

Each release produces:
```
agent-deck_X.Y.Z_darwin_amd64.tar.gz   # macOS Intel
agent-deck_X.Y.Z_darwin_arm64.tar.gz   # macOS Apple Silicon
agent-deck_X.Y.Z_linux_amd64.tar.gz    # Linux x86_64
agent-deck_X.Y.Z_linux_arm64.tar.gz    # Linux ARM64
checksums.txt                           # SHA256 checksums
```

### Required GitHub Secrets

| Secret | Purpose |
|--------|---------|
| `GITHUB_TOKEN` | Auto-provided, creates releases |
| `HOMEBREW_TAP_GITHUB_TOKEN` | Push to homebrew-tap repo |

### Local Testing

```bash
# Test release locally (no publish)
goreleaser release --snapshot --clean

# Check config
goreleaser check
```

## Colors (Tokyo Night)

| Color | Hex | Use |
|-------|-----|-----|
| Accent | `#7aa2f7` | Selection, highlights |
| Green | `#9ece6a` | Running status |
| Yellow | `#e0af68` | Waiting status |
| Red | `#f7768e` | Error status |
| Cyan | `#7dcfff` | Group names |
| Purple | `#bb9af7` | Tool tags |
| Background | `#1a1b26` | Dark background |
| Surface | `#24283b` | Panel backgrounds |

## Key Behaviors

- **Polling**: Status checked every 500ms via tick loop
- **No yellow flash**: Sessions start as idle (gray) on init; busy indicator check catches active ones
- **Busy detection**: "esc to interrupt", spinners → immediate GREEN (before hash comparison)
- **Persistence**: Status survives restart (acknowledged state stored in JSON)
- **Isolation**: Each session has its own StateTracker
- **Lazy loading**: Sessions loaded on app start, `UpdateStatus()` called immediately
- **Atomic writes**: Single JSON file, no concurrent writers

## Debugging

### Status Detection Debug Mode
```bash
AGENTDECK_DEBUG=1 agent-deck
```

Logs status transitions to stderr:
```
[STATUS] my-project: BUSY INDICATOR → active
[STATUS] my-project: CHANGED → active
[STATUS] my-project: COOLDOWN → active
[STATUS] my-project: WAITING → waiting
[STATUS] my-project: IDLE → idle
```

### Common Issues
| Symptom | Cause | Fix |
|---------|-------|-----|
| Yellow when should be green | Busy indicator not detected | Check `hasBusyIndicator()` patterns |
| Yellow flash on startup | Init returning "waiting" | Should return "idle", busy check catches active |
| Status not persisting | `acknowledged` not saved | Check `ReconnectSessionWithStatus()` |
| Flickering green/yellow | Cooldown too short | Increase `activityCooldown` (default 2s) |
