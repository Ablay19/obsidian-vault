package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Commit represents a Git commit with metadata and analysis results
type Commit struct {
	ID                 string    `json:"id" db:"id"`
	Hash               string    `json:"hash" db:"hash"`
	BranchID           string    `json:"branch_id" db:"branch_id"`
	AuthorName         string    `json:"author_name" db:"author_name"`
	AuthorEmail        string    `json:"author_email" db:"author_email"`
	CommitDate         time.Time `json:"commit_date" db:"commit_date"`
	Message            string    `json:"message" db:"message"`
	ParentHashes       []string  `json:"parent_hashes" db:"parent_hashes"`
	ChangedFiles       []string  `json:"changed_files" db:"changed_files"`
	Insertions         int       `json:"insertions" db:"insertions"`
	Deletions          int       `json:"deletions" db:"deletions"`
	QualityScore       int       `json:"quality_score" db:"quality_score"`
	HasSecurityIssues  bool      `json:"has_security_issues" db:"has_security_issues"`
	IsLargeCommit      bool      `json:"is_large_commit" db:"is_large_commit"`
	SensitiveDataFound bool      `json:"sensitive_data_found" db:"sensitive_data_found"`
	CreatedAt          time.Time `json:"created_at" db:"created_at"`
}

// NewCommit creates a new Commit with default values
func NewCommit(hash, branchID, authorName, authorEmail string, commitDate time.Time, message string) *Commit {
	now := time.Now()
	return &Commit{
		ID:                 uuid.New().String(),
		Hash:               hash,
		BranchID:           branchID,
		AuthorName:         authorName,
		AuthorEmail:        authorEmail,
		CommitDate:         commitDate,
		Message:            message,
		ParentHashes:       []string{},
		ChangedFiles:       []string{},
		Insertions:         0,
		Deletions:          0,
		QualityScore:       0,
		HasSecurityIssues:  false,
		IsLargeCommit:      false,
		SensitiveDataFound: false,
		CreatedAt:          now,
	}
}

// Validate checks if the commit data is valid
func (c *Commit) Validate() error {
	if c.Hash == "" {
		return errors.New("commit hash cannot be empty")
	}
	if c.BranchID == "" {
		return errors.New("branch ID cannot be empty")
	}
	if c.AuthorName == "" {
		return errors.New("author name cannot be empty")
	}
	if c.AuthorEmail == "" {
		return errors.New("author email cannot be empty")
	}
	if c.Message == "" {
		return errors.New("commit message cannot be empty")
	}
	if len(c.Message) > 500 {
		return errors.New("commit message too long")
	}
	if c.QualityScore < 0 || c.QualityScore > 100 {
		return errors.New("quality score must be between 0 and 100")
	}
	return nil
}

// TotalChanges returns the total number of lines changed
func (c *Commit) TotalChanges() int {
	return c.Insertions + c.Deletions
}
