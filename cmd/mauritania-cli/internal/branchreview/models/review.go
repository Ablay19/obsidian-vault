package models

import (
	"errors"
	"time"

	"github.com/google/uuid"
)

// Review represents a review session with findings and recommendations
type Review struct {
	ID               string       `json:"id" db:"id"`
	TargetType       ReviewTarget `json:"target_type" db:"target_type"`
	TargetIdentifier string       `json:"target_identifier" db:"target_identifier"`
	ReviewType       ReviewType   `json:"review_type" db:"review_type"`
	Status           ReviewStatus `json:"status" db:"status"`
	StartedAt        *time.Time   `json:"started_at,omitempty" db:"started_at"`
	CompletedAt      *time.Time   `json:"completed_at,omitempty" db:"completed_at"`
	TotalCommits     int          `json:"total_commits" db:"total_commits"`
	IssuesFound      int          `json:"issues_found" db:"issues_found"`
	CriticalIssues   int          `json:"critical_issues" db:"critical_issues"`
	Recommendations  []string     `json:"recommendations" db:"recommendations"`
	ReportURL        string       `json:"report_url,omitempty" db:"report_url"`
	Reviewer         string       `json:"reviewer,omitempty" db:"reviewer"`
	CreatedAt        time.Time    `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time    `json:"updated_at" db:"updated_at"`
}

// ReviewTarget represents the type of target being reviewed
type ReviewTarget string

const (
	ReviewTargetBranch      ReviewTarget = "branch"
	ReviewTargetCommitRange ReviewTarget = "commit_range"
	ReviewTargetRepository  ReviewTarget = "repository"
)

// ReviewType represents the type of review being performed
type ReviewType string

const (
	ReviewTypeIndividual    ReviewType = "individual"
	ReviewTypeComprehensive ReviewType = "comprehensive"
	ReviewTypeSecurity      ReviewType = "security"
)

// ReviewStatus represents the status of a review
type ReviewStatus string

const (
	ReviewStatusPending    ReviewStatus = "pending"
	ReviewStatusInProgress ReviewStatus = "in_progress"
	ReviewStatusCompleted  ReviewStatus = "completed"
	ReviewStatusFailed     ReviewStatus = "failed"
)

// NewReview creates a new Review with default values
func NewReview(targetType ReviewTarget, targetIdentifier string, reviewType ReviewType, reviewer string) *Review {
	now := time.Now()
	return &Review{
		ID:               uuid.New().String(),
		TargetType:       targetType,
		TargetIdentifier: targetIdentifier,
		ReviewType:       reviewType,
		Status:           ReviewStatusPending,
		TotalCommits:     0,
		IssuesFound:      0,
		CriticalIssues:   0,
		Recommendations:  []string{},
		Reviewer:         reviewer,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
}

// Validate checks if the review data is valid
func (r *Review) Validate() error {
	if r.TargetIdentifier == "" {
		return errors.New("target identifier cannot be empty")
	}
	if r.IssuesFound < 0 {
		return errors.New("issues found cannot be negative")
	}
	if r.CriticalIssues < 0 {
		return errors.New("critical issues cannot be negative")
	}
	if len(r.Recommendations) > 50 {
		return errors.New("too many recommendations")
	}
	return nil
}

// Start marks the review as in progress
func (r *Review) Start() {
	now := time.Now()
	r.Status = ReviewStatusInProgress
	r.StartedAt = &now
	r.UpdatedAt = now
}

// Complete marks the review as completed
func (r *Review) Complete() {
	now := time.Now()
	r.Status = ReviewStatusCompleted
	r.CompletedAt = &now
	r.UpdatedAt = now
}

// Fail marks the review as failed
func (r *Review) Fail() {
	now := time.Now()
	r.Status = ReviewStatusFailed
	r.CompletedAt = &now
	r.UpdatedAt = now
}
