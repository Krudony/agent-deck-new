package session

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"
)

// CodexSessionInfo contains information about a Codex session
type CodexSessionInfo struct {
	SessionID   string
	Filename    string
	LastUpdated time.Time
}

// GetCodexSessionsDir returns the Codex sessions directory
func GetCodexSessionsDir() string {
	home, _ := os.UserHomeDir()
	return filepath.Join(home, ".codex", "sessions")
}

// ListCodexSessions returns all Codex sessions
// Scans ~/.codex/sessions/YYYY/MM/DD/*.jsonl files
// Sorted by LastUpdated (most recent first)
func ListCodexSessions() ([]CodexSessionInfo, error) {
	sessionsDir := GetCodexSessionsDir()

	// Find all .jsonl session files recursively
	var files []string
	err := filepath.Walk(sessionsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // Skip errors
		}
		if !info.IsDir() && strings.HasSuffix(info.Name(), ".jsonl") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}

	// Extract session ID from filename: rollout-{timestamp}-{UUID}.jsonl
	// UUID format: 019b914d-17d3-7110-b151-6158833bf32e
	sessionIDPattern := regexp.MustCompile(`rollout-\d{4}-\d{2}-\d{2}T\d{2}-\d{2}-\d{2}-([0-9a-f\-]+)\.jsonl`)

	var sessions []CodexSessionInfo
	for _, file := range files {
		filename := filepath.Base(file)
		matches := sessionIDPattern.FindStringSubmatch(filename)
		if len(matches) != 2 {
			continue // Skip files that don't match pattern
		}

		sessionID := matches[1]
		info, err := os.Stat(file)
		if err != nil {
			continue
		}

		sessions = append(sessions, CodexSessionInfo{
			SessionID:   sessionID,
			Filename:    filename,
			LastUpdated: info.ModTime(),
		})
	}

	// Sort by LastUpdated (most recent first)
	sort.Slice(sessions, func(i, j int) bool {
		return sessions[i].LastUpdated.After(sessions[j].LastUpdated)
	})

	return sessions, nil
}
