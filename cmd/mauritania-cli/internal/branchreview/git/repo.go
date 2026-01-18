package git

import (
	"context"

	"github.com/go-git/go-git/v5/plumbing/object"
)

// Repository defines the interface for Git repository operations
type Repository interface {
	// Open opens a Git repository at the given path
	Open(ctx context.Context, path string) error

	// GetBranches returns all branches in the repository
	GetBranches(ctx context.Context) ([]BranchInfo, error)

	// GetCommits returns commits for a specific branch
	GetCommits(ctx context.Context, branchName string, limit int) ([]CommitInfo, error)

	// GetCommit returns a specific commit by hash
	GetCommit(ctx context.Context, hash string) (*CommitInfo, error)

	// Close closes the repository connection
	Close() error
}

// BranchInfo represents basic branch information
type BranchInfo struct {
	Name       string
	Hash       string
	IsRemote   bool
	LastCommit *CommitInfo
}

// CommitInfo represents commit information
type CommitInfo struct {
	Hash         string
	AuthorName   string
	AuthorEmail  string
	Message      string
	Date         object.Signature
	Parents      []string
	ChangedFiles []string
	Insertions   int
	Deletions    int
}
