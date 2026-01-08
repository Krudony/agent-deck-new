# Custom Features Requirements

> **Purpose**: This document lists all custom features added to agent-deck that are NOT in the original upstream repository. Use this as a checklist when merging upstream updates.

---

## 📋 Overview

**Base Version**: v0.8.10 (upstream)
**Custom Version**: v1.0.0
**Total Changes**: 7 files modified, 210 insertions(+), 76 deletions(-)

---

## 🎯 Feature 1: Codex Session Resume Support

### Summary
Enable R key to resume Codex CLI sessions with full session persistence and tracking.

### Files Changed

#### 1. `internal/session/codex.go` (NEW FILE - 77 lines)
**Purpose**: Codex session detection and management

```go
// Required structs
type CodexSessionInfo struct {
    SessionID   string
    Filename    string
    LastUpdated time.Time
}

// Required functions
func GetCodexSessionsDir() string
func ListCodexSessions() ([]CodexSessionInfo, error)
```

**Implementation Details**:
- Scan `~/.codex/sessions/YYYY/MM/DD/*.jsonl` recursively
- Extract session ID from filename pattern: `rollout-{timestamp}-{UUID}.jsonl`
- Sort by LastUpdated (most recent first)
- UUID format: `019b914d-17d3-7110-b151-6158833bf32e`

---

#### 2. `internal/session/instance.go` (157 lines changed)

**A. Add Codex fields to Instance struct** (line ~54-56)
```go
// Codex CLI integration
CodexSessionID  string    `json:"codex_session_id,omitempty"`
CodexDetectedAt time.Time `json:"codex_detected_at,omitempty"`
```

**B. Add UpdateCodexSession() function** (line ~557-580)
```go
// UpdateCodexSession updates the Codex session ID from files.
// Scans ~/.codex/sessions/ for the most recent session file.
func (i *Instance) UpdateCodexSession(excludeIDs map[string]bool) {
    if i.Tool != "codex" {
        return
    }

    // Always scan for the most recent session
    sessions, err := ListCodexSessions()
    if err != nil || len(sessions) == 0 {
        return
    }

    // Use the most recent session (already sorted by LastUpdated)
    for _, sess := range sessions {
        // Skip excluded IDs
        if excludeIDs != nil && excludeIDs[sess.SessionID] {
            continue
        }
        i.CodexSessionID = sess.SessionID
        i.CodexDetectedAt = time.Now()
        return
    }
}
```

**C. Update Start() method to call UpdateCodexSession()** (line ~501-504)
```go
// Update Codex session tracking (non-blocking, best-effort)
if i.Tool == "codex" {
    i.UpdateCodexSession(nil)
}
```

**D. Update Restart() method log** (line ~1037-1038)
```go
log.Printf("[MCP-DEBUG] Instance.Restart() called - Tool=%s, ClaudeSessionID=%q, GeminiSessionID=%q, CodexSessionID=%q, tmuxSession=%v, tmuxExists=%v",
    i.Tool, i.ClaudeSessionID, i.GeminiSessionID, i.CodexSessionID, i.tmuxSession != nil, i.tmuxSession != nil && i.tmuxSession.Exists())
```

**E. Add Codex respawn-pane support in Restart()** (line ~1085-1098)
```go
// If Codex session with known ID AND tmux session exists, use respawn-pane
if i.Tool == "codex" && i.CodexSessionID != "" && i.tmuxSession != nil && i.tmuxSession.Exists() {
    resumeCmd := fmt.Sprintf("codex resume %s", i.CodexSessionID)
    log.Printf("[MCP-DEBUG] Codex respawn-pane with command: %s", resumeCmd)

    if err := i.tmuxSession.RespawnPane(resumeCmd); err != nil {
        log.Printf("[MCP-DEBUG] Codex RespawnPane failed: %v", err)
        return fmt.Errorf("failed to restart Codex session: %w", err)
    }

    log.Printf("[MCP-DEBUG] Codex RespawnPane succeeded")
    i.Status = StatusWaiting
    return nil
}
```

**F. Add Codex fallback in Restart()** (line ~1112-1113)
```go
} else if i.Tool == "codex" && i.CodexSessionID != "" {
    command = fmt.Sprintf("codex resume %s", i.CodexSessionID)
```

**G. Add Codex support in CanRestart()** (line ~1171-1174)
```go
// Codex sessions with known session ID can always be restarted
if i.Tool == "codex" && i.CodexSessionID != "" {
    return true
}
```

---

## 🎯 Feature 2: Claude Dangerous Mode Default to True

### Summary
Change Claude dangerous mode to default to `true` (enabled) instead of `false`. Use pointer type to distinguish "not set" from "explicitly false".

### Files Changed

#### 1. `internal/session/userconfig.go` (4 lines changed)

**Change DangerousMode type** (line ~123-125)
```go
// BEFORE
DangerousMode bool `toml:"dangerous_mode"`

// AFTER
// DangerousMode enables --dangerously-skip-permissions flag for Claude sessions
// Default: true (nil = use default true, explicitly set false to disable)
DangerousMode *bool `toml:"dangerous_mode"`
```

---

#### 2. `internal/session/instance.go` (3 locations)

**Location A: buildClaudeCommandWithMessage()** (line ~192-198)
```go
// BEFORE
dangerousMode := false
if userConfig, err := LoadUserConfig(); err == nil && userConfig != nil && !userConfig.Claude.DangerousMode {
    dangerousMode = userConfig.Claude.DangerousMode
}

// AFTER
// Check if dangerous mode is enabled in user config
// Default to true (always use --dangerously-skip-permissions)
dangerousMode := true
if userConfig, err := LoadUserConfig(); err == nil && userConfig != nil && userConfig.Claude.DangerousMode != nil {
    // Use explicit value from config (nil = use default true)
    dangerousMode = *userConfig.Claude.DangerousMode
}
```

**Location B: buildClaudeResumeCommand()** (line ~1223-1226)
```go
// Same pattern as Location A
dangerousMode := true
if userConfig, err := LoadUserConfig(); err == nil && userConfig != nil && userConfig.Claude.DangerousMode != nil {
    dangerousMode = *userConfig.Claude.DangerousMode
}
```

**Location C: Fork()** (line ~1223-1226)
```go
// Same pattern as Location A and B
dangerousMode := true
if userConfig, err := LoadUserConfig(); err == nil && userConfig != nil && userConfig.Claude.DangerousMode != nil {
    dangerousMode = *userConfig.Claude.DangerousMode
}
```

---

#### 3. `internal/ui/home.go` (6 lines changed)

**Update dangerous mode logic** (line ~1689-1696)
```go
// BEFORE
dangerousMode := false
configDir := ""
if userConfig != nil {
    dangerousMode = userConfig.Claude.DangerousMode
    configDir = userConfig.Claude.ConfigDir
}

// AFTER
dangerousMode := true
configDir := ""
if userConfig != nil {
    if userConfig.Claude.DangerousMode != nil {
        dangerousMode = *userConfig.Claude.DangerousMode
    }
    configDir = userConfig.Claude.ConfigDir
}
```

---

#### 4. `internal/session/instance_test.go` (38 lines changed)

**Test: TestInstance_Fork_RespectsDangerousMode** (line ~827-832, 855-860)

```go
// BEFORE
t.Run("dangerous_mode=false", func(t *testing.T) {
    userConfigCache = &UserConfig{
        Claude: ClaudeSettings{
            DangerousMode: false,
        },
    }
})

// AFTER
t.Run("dangerous_mode=false", func(t *testing.T) {
    falseBool := false
    userConfigCache = &UserConfig{
        Claude: ClaudeSettings{
            DangerousMode: &falseBool,
        },
    }
})

// Same for dangerous_mode=true test:
trueBool := true
userConfigCache = &UserConfig{
    Claude: ClaudeSettings{
        DangerousMode: &trueBool,
    },
}
```

---

## 🎯 Feature 3: Gemini Session Resume Fix (R key)

### Summary
Fix Gemini R key not working - always scan for newest session instead of returning early.

### Files Changed

#### 1. `internal/session/instance.go` (29 lines changed)

**UpdateGeminiSession() function** (line ~529-555)

```go
// BEFORE
func (i *Instance) UpdateGeminiSession(excludeIDs map[string]bool) {
    if i.Tool != "gemini" {
        return
    }

    // If we already have a session ID, check if tmux env confirms it
    if i.GeminiSessionID != "" {
        // Only update timestamp if we can confirm from tmux env
        if i.tmuxSession != nil && i.tmuxSession.Exists() {
            if envID, err := i.tmuxSession.GetEnvironment("GEMINI_SESSION_ID"); err == nil && envID == i.GeminiSessionID {
                i.GeminiDetectedAt = time.Now()
            }
        }
        return  // <-- PROBLEM: Early return prevents detecting new sessions
    }

    // Scan for most recent session from files
    sessions, err := ListGeminiSessions(i.ProjectPath)
    // ... rest of function
}

// AFTER
func (i *Instance) UpdateGeminiSession(excludeIDs map[string]bool) {
    if i.Tool != "gemini" {
        return
    }

    // Always scan for the most recent session to handle cases where:
    // - User started a new session but we still have old ID
    // - Session ID changed but tmux env wasn't updated
    // This ensures we always use the CURRENT session

    // Scan for most recent session from files
    sessions, err := ListGeminiSessions(i.ProjectPath)
    if err != nil || len(sessions) == 0 {
        return
    }

    // Use the most recent session (already sorted by LastUpdated)
    for _, sess := range sessions {
        // Skip excluded IDs
        if excludeIDs != nil && excludeIDs[sess.SessionID] {
            continue
        }
        i.GeminiSessionID = sess.SessionID
        i.GeminiDetectedAt = time.Now()
        return
    }
}
```

**Key Change**: Remove early return when `i.GeminiSessionID != ""` to always scan for newest session.

---

## 🎯 Feature 4: Repository References Update

### Summary
Update repository URLs to point to agent-deck-new

### Files Changed

#### 1. `install.sh` (2 lines changed)
```bash
# BEFORE
REPO="Krudony/Krudonagent-deck"

# AFTER
REPO="Krudony/agent-deck-new"
```

#### 2. `internal/update/update.go` (2 lines changed)
```go
// BEFORE
GitHubRepo = "Krudony/Krudonagent-deck"

// AFTER
GitHubRepo = "Krudony/agent-deck-new"
```

---

## ✅ Verification Checklist

After applying these features to a new upstream version, verify:

- [ ] `internal/session/codex.go` exists with 77 lines
- [ ] `ListCodexSessions()` function works (test: `go test ./internal/session/...`)
- [ ] Codex fields added to Instance struct (CodexSessionID, CodexDetectedAt)
- [ ] UpdateCodexSession() function exists and is called in Start()
- [ ] Codex support in CanRestart() returns true when CodexSessionID != ""
- [ ] Codex support in Restart() uses respawn-pane
- [ ] DangerousMode is `*bool` type in userconfig.go
- [ ] All dangerous mode checks use `dangerousMode := true` as default
- [ ] All dangerous mode checks use `*userConfig.Claude.DangerousMode` (pointer dereference)
- [ ] Tests use `falseBool := false; DangerousMode: &falseBool` pattern
- [ ] UpdateGeminiSession() always scans (no early return)
- [ ] install.sh points to correct repo
- [ ] internal/update/update.go points to correct repo
- [ ] All tests pass: `go test ./...`
- [ ] Binary builds: `go build -o build/agent-deck ./cmd/agent-deck`

---

## 📊 Statistics

```
Total files changed: 7
- New files: 1 (codex.go)
- Modified files: 6

Lines changed:
- Insertions: +210
- Deletions: -76

Key files:
- instance.go: 157 lines changed (13 hunks)
- codex.go: 77 lines (new)
- instance_test.go: 38 lines changed
- userconfig.go: 4 lines changed
- home.go: 6 lines changed
- install.sh: 2 lines changed
- update.go: 2 lines changed
```

---

## 🔄 How to Apply These Features

### Quick Method (If using Git)
```bash
# 1. Update to latest upstream
git pull upstream main

# 2. Cherry-pick the custom features commit
git cherry-pick <commit-hash-of-custom-features>

# 3. Resolve any conflicts
git status
# Fix conflicts in marked files
git add .
git cherry-pick --continue

# 4. Build and test
go build -o build/agent-deck ./cmd/agent-deck
go test ./...
```

### Manual Method
1. Copy `internal/session/codex.go` from this repo
2. Apply changes to `internal/session/instance.go` following the line numbers above
3. Apply changes to `internal/session/userconfig.go`
4. Apply changes to `internal/session/instance_test.go`
5. Apply changes to `internal/ui/home.go`
6. Update `install.sh` and `internal/update/update.go`
7. Build and test

---

## 📝 Notes

- **Codex Support**: Requires Codex CLI to be installed separately
- **Dangerous Mode**: Default is `true` (enabled). Users can disable by setting `dangerous_mode = false` in config
- **Gemini Fix**: Critical for R key to work on machines with multiple Gemini sessions
- **Pointer Type**: Using `*bool` allows distinguishing between "not set" (nil) vs "explicitly false"

---

**Last Updated**: 2026-01-08
**Maintained By**: Krudony
**Based On**: agent-deck v0.8.10 (upstream)
