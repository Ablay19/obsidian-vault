package git

import (
	"fmt"
	"log/slog"
	"os/exec"
	"strings"
	"time"
)

// Manager handles Git operations for a specific repository path.
type Manager struct {
	RepoPath string
}

// NewManager creates a new Git manager.
func NewManager(repoPath string) *Manager {
	return &Manager{RepoPath: repoPath}
}

// RunCommand executes a git command in the repository path.
func (m *Manager) RunCommand(args ...string) (string, error) {
	cmd := exec.Command("git", append([]string{"-C", m.RepoPath}, args...)...)
	out, err := cmd.CombinedOutput()
	output := string(out)
	if err != nil {
		return output, fmt.Errorf("git %s failed: %w (output: %s)", strings.Join(args, " "), err, output)
	}
	return output, nil
}

// ConfigureUser sets local git user config.
func (m *Manager) ConfigureUser(name, email string) error {
	slog.Info("Configuring Git user", "name", name, "email", email)
	if _, err := m.RunCommand("config", "--local", "user.name", name); err != nil {
		return err
	}
	if _, err := m.RunCommand("config", "--local", "user.email", email); err != nil {
		return err
	}
	return nil
}

// SyncAutoCommit adds, commits and pushes changes.
func (m *Manager) SyncAutoCommit(message string) error {
	slog.Info("Starting Git sync", "message", message)

	// 1. Add
	if _, err := m.RunCommand("add", "."); err != nil {
		return err
	}

	// Check if there are changes to commit
	status, err := m.RunCommand("status", "--porcelain")
	if err != nil {
		return err
	}
	if strings.TrimSpace(status) == "" {
		slog.Info("No changes to commit")
		return nil
	}

	// 2. Commit
	if _, err := m.RunCommand("commit", "-m", message); err != nil {
		return err
	}

	// 3. Push with retries and force-with-lease if needed
	return m.PushWithRetry(3)
}

// PushWithRetry attempts to push changes multiple times.
func (m *Manager) PushWithRetry(maxRetries int) error {
	branch, err := m.GetCurrentBranch()
	if err != nil {
		branch = "main" // Default fallback
	}

	for i := 0; i < maxRetries; i++ {
		slog.Info("Attempting Git push", "branch", branch, "attempt", i+1)
		
		// Try standard push first
		_, err := m.RunCommand("push", "origin", branch)
		if err == nil {
			slog.Info("Git push successful")
			return nil
		}

		slog.Warn("Git push failed, attempting pull/rebase", "error", err)
		
		// If push failed, try to pull/rebase to resolve potential conflicts
		if _, err := m.RunCommand("pull", "--rebase", "origin", branch); err != nil {
			slog.Error("Git pull --rebase failed", "error", err)
			// If rebase fails, we might need to abort it
			m.RunCommand("rebase", "--abort")
		} else {
			// Pull successful, retry push in next loop
			continue
		}

		// As a last resort, try force-with-lease on the last attempt
		if i == maxRetries-1 {
			slog.Warn("Attempting force-with-lease push as last resort")
			_, err = m.RunCommand("push", "--force-with-lease", "origin", branch)
			return err
		}

		time.Sleep(2 * time.Second)
	}

	return fmt.Errorf("git push failed after %d attempts", maxRetries)
}

// GetCurrentBranch returns the name of the current branch.
func (m *Manager) GetCurrentBranch() (string, error) {
	out, err := m.RunCommand("rev-parse", "--abbrev-ref", "HEAD")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

// EnsureRemote ensures the origin remote is set.
func (m *Manager) EnsureRemote(url string) error {
	_, err := m.RunCommand("remote", "get-url", "origin")
	if err != nil {
		slog.Info("Adding missing origin remote", "url", url)
		_, err = m.RunCommand("remote", "add", "origin", url)
		return err
	}
	return nil
}
