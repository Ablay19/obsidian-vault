package models

import (
	"time"

	"github.com/google/uuid"
)

// Issue represents a specific problem or finding from analysis
type Issue struct {
	ID          string        `json:"id" db:"id"`
	ReviewID    string        `json:"review_id" db:"review_id"`
	CommitHash  string        `json:"commit_hash" db:"commit_hash"`
	IssueType   IssueType     `json:"issue_type" db:"issue_type"`
	Severity    IssueSeverity `json:"severity" db:"severity"`
	Category    string        `json:"category" db:"category"`
	Description string        `json:"description" db:"description"`
	FilePath    string        `json:"file_path,omitempty" db:"file_path"`
	LineNumber  int           `json:"line_number,omitempty" db:"line_number"`
	CodeSnippet string        `json:"code_snippet,omitempty" db:"code_snippet"`
	Remediation string        `json:"remediation,omitempty" db:"remediation"`
	ToolName    string        `json:"tool_name,omitempty" db:"tool_name"`
	Confidence  float64       `json:"confidence" db:"confidence"`
	CreatedAt   time.Time     `json:"created_at" db:"created_at"`
}

// IssueType represents the type of issue found
type IssueType string

const (
	IssueTypeSecurity      IssueType = "security"
	IssueTypeQuality       IssueType = "quality"
	IssueTypePerformance   IssueType = "performance"
	IssueTypeDocumentation IssueType = "documentation"
)

// IssueSeverity represents the severity level of an issue
type IssueSeverity string

const (
	IssueSeverityCritical IssueSeverity = "critical"
	IssueSeverityHigh     IssueSeverity = "high"
	IssueSeverityMedium   IssueSeverity = "medium"
	IssueSeverityLow      IssueSeverity = "low"
	IssueSeverityInfo     IssueSeverity = "info"
)

// NewIssue creates a new Issue with default values
func NewIssue(reviewID, commitHash string, issueType IssueType, severity IssueSeverity, description string) *Issue {
	now := time.Now()
	return &Issue{
		ID:          uuid.New().String(),
		ReviewID:    reviewID,
		CommitHash:  commitHash,
		IssueType:   issueType,
		Severity:    severity,
		Description: description,
		Confidence:  0.0,
		CreatedAt:   now,
	}
}

// Validate checks if the issue data is valid
func (i *Issue) Validate() error {
	if i.ReviewID == "" {
		return errors.New("review ID cannot be empty")
	}
	if i.CommitHash == "" {
		return errors.New("commit hash cannot be empty")
	}
	if i.Description == "" {
		return errors.New("description cannot be empty")
	}
	if len(i.Description) > 1000 {
		return errors.New("description too long")
	}
	if i.LineNumber < 0 {
		return errors.New("line number cannot be negative")
	}
	if len(i.CodeSnippet) > 500 {
		return errors.New("code snippet too long")
	}
	if i.Confidence < 0.0 || i.Confidence > 1.0 {
		return errors.New("confidence must be between 0.0 and 1.0")
	}
	return nil
}
