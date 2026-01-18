package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Branch represents a Git branch with its current state and metadata
type Branch struct {
	ID                string       `json:"id" db:"id"`
	Name              string       `json:"name" db:"name"`
	RepositoryURL     string       `json:"repository_url" db:"repository_url"`
	Status            BranchStatus `json:"status" db:"status"`
	LastCommitHash    string       `json:"last_commit_hash,omitempty" db:"last_commit_hash"`
	LastCommitDate    *time.Time   `json:"last_commit_date,omitempty" db:"last_commit_date"`
	LastCommitAuthor  string       `json:"last_commit_author,omitempty" db:"last_commit_author"`
	AheadCount        int          `json:"ahead_count" db:"ahead_count"`
	BehindCount       int          `json:"behind_count" db:"behind_count"`
	HasMergeConflicts bool         `json:"has_merge_conflicts" db:"has_merge_conflicts"`
	CreatedAt         time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt         time.Time    `json:"updated_at" db:"updated_at"`
}

// BranchStatus represents the status of a branch
type BranchStatus string

const (
	BranchStatusActive  BranchStatus = "active"
	BranchStatusStale   BranchStatus = "stale"
	BranchStatusMerged  BranchStatus = "merged"
	BranchStatusDeleted BranchStatus = "deleted"
)

// NewBranch creates a new Branch with default values
func NewBranch(name, repoURL string) *Branch {
	now := time.Now()
	return &Branch{
		ID:            uuid.New().String(),
		Name:          name,
		RepositoryURL: repoURL,
		Status:        BranchStatusActive,
		AheadCount:    0,
		BehindCount:   0,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
}

// Validate checks if the branch data is valid
func (b *Branch) Validate() error {
	if b.Name == "" {
		return errors.New("invalid branch name")
	}
	if b.RepositoryURL == "" {
		return errors.New("invalid repository URL")
	}
	return nil
}
