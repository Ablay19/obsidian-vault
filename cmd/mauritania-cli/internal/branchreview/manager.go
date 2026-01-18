package branchreview

import (
	"context"
	"fmt"

	"obsidian-automation/cmd/mauritania-cli/internal/branchreview/git"
	"obsidian-automation/cmd/mauritania-cli/internal/branchreview/models"
)

// RepositoryManager manages Git repository operations
type RepositoryManager struct {
	repo git.Repository
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(repo git.Repository) *RepositoryManager {
	return &RepositoryManager{
		repo: repo,
	}
}

// OpenRepository opens a repository at the given path
func (rm *RepositoryManager) OpenRepository(ctx context.Context, path string) error {
	return rm.repo.Open(ctx, path)
}

// GetBranchInfo gets information about a specific branch
func (rm *RepositoryManager) GetBranchInfo(ctx context.Context, branchName string) (*models.Branch, error) {
	branches, err := git.ListBranches(ctx, rm.repo)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}

	for _, branch := range branches {
		if branch.Name == branchName {
			model := &models.Branch{
				Name:           branch.Name,
				RepositoryURL:  "local", // TODO: get actual repo URL
				Status:         models.BranchStatusActive,
				LastCommitHash: branch.Hash,
			}

			if branch.LastCommit != nil {
				model.LastCommitAuthor = branch.LastCommit.AuthorName
			}

			return model, nil
		}
	}

	return nil, fmt.Errorf("branch %s not found", branchName)
}

// GetBranchCommits gets commits for a branch and converts to models
func (rm *RepositoryManager) GetBranchCommits(ctx context.Context, branchName string, limit int) ([]*models.Commit, error) {
	commits, err := git.ListCommits(ctx, rm.repo, branchName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits: %w", err)
	}

	models := make([]*models.Commit, len(commits))
	for i, commit := range commits {
		models[i] = &models.Commit{
			Hash:         commit.Hash,
			BranchID:     branchName, // TODO: use proper branch ID
			AuthorName:   commit.AuthorName,
			AuthorEmail:  commit.AuthorEmail,
			CommitDate:   commit.Date.When,
			Message:      commit.Message,
			ParentHashes: commit.Parents,
			ChangedFiles: commit.ChangedFiles,
			Insertions:   commit.Insertions,
			Deletions:    commit.Deletions,
			QualityScore: 0, // TODO: calculate quality score
		}

		// Mark large commits
		if models[i].TotalChanges() > 1000 {
			models[i].IsLargeCommit = true
		}
	}

	return models, nil
}

// Close closes the repository
func (rm *RepositoryManager) Close() error {
	return rm.repo.Close()
}
