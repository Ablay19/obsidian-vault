package git

import (
	"context"
	"fmt"
)

// ListBranches lists all branches in the repository
func ListBranches(ctx context.Context, repo Repository) ([]BranchInfo, error) {
	branches, err := repo.GetBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list branches: %w", err)
	}
	return branches, nil
}

// GetBranchByName finds a branch by name
func GetBranchByName(ctx context.Context, repo Repository, name string) (*BranchInfo, error) {
	branches, err := repo.GetBranches(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	for _, branch := range branches {
		if branch.Name == name {
			return &branch, nil
		}
	}

	return nil, fmt.Errorf("branch %s not found", name)
}

// FilterActiveBranches returns only active branches (not stale)
func FilterActiveBranches(branches []BranchInfo) []BranchInfo {
	active := []BranchInfo{}
	for _, branch := range branches {
		// For now, consider all branches active
		// In future, implement logic to detect stale branches
		active = append(active, branch)
	}
	return active
}
