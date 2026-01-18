package git

import (
	"context"
	"fmt"
)

// ListCommits lists commits for a branch
func ListCommits(ctx context.Context, repo Repository, branchName string, limit int) ([]CommitInfo, error) {
	commits, err := repo.GetCommits(ctx, branchName, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list commits for branch %s: %w", branchName, err)
	}
	return commits, nil
}

// GetLatestCommit gets the most recent commit for a branch
func GetLatestCommit(ctx context.Context, repo Repository, branchName string) (*CommitInfo, error) {
	commits, err := repo.GetCommits(ctx, branchName, 1)
	if err != nil {
		return nil, fmt.Errorf("failed to get latest commit for branch %s: %w", branchName, err)
	}
	if len(commits) == 0 {
		return nil, fmt.Errorf("no commits found for branch %s", branchName)
	}
	return &commits[0], nil
}

// FilterLargeCommits returns commits with more than the specified number of changes
func FilterLargeCommits(commits []CommitInfo, minChanges int) []CommitInfo {
	large := []CommitInfo{}
	for _, commit := range commits {
		if commit.Insertions+commit.Deletions > minChanges {
			large = append(large, commit)
		}
	}
	return large
}

// CountCommitsByAuthor counts commits by author
func CountCommitsByAuthor(commits []CommitInfo) map[string]int {
	counts := make(map[string]int)
	for _, commit := range commits {
		counts[commit.AuthorEmail]++
	}
	return counts
}
