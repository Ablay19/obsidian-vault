package e2e

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"obsidian-automation/cmd/mauritania-cli/internal/branchreview/git"
	"obsidian-automation/cmd/mauritania-cli/internal/branchreview/models"
)

func TestGitOperations(t *testing.T) {
	ctx := context.Background()

	// Use local repository
	repo := git.NewLocalRepository()
	defer repo.Close()

	// Open current repository
	err := repo.Open(ctx, ".")
	require.NoError(t, err)

	// Test listing branches
	branches, err := repo.GetBranches(ctx)
	require.NoError(t, err)
	assert.NotEmpty(t, branches)

	// Find main or master branch
	var mainBranch *git.BranchInfo
	for _, branch := range branches {
		if branch.Name == "main" || branch.Name == "master" {
			mainBranch = &branch
			break
		}
	}
	require.NotNil(t, mainBranch, "main or master branch not found")

	// Test getting commits
	commits, err := repo.GetCommits(ctx, mainBranch.Name, 5)
	require.NoError(t, err)
	assert.NotEmpty(t, commits)

	// Test converting to models
	branchModel := &models.Branch{
		Name:          mainBranch.Name,
		RepositoryURL: "local",
		Status:        models.BranchStatusActive,
	}

	if mainBranch.LastCommit != nil {
		branchModel.LastCommitHash = mainBranch.LastCommit.Hash
		branchModel.LastCommitAuthor = mainBranch.LastCommit.AuthorName
	}

	err = branchModel.Validate()
	assert.NoError(t, err)
}

func TestBranchInfo(t *testing.T) {
	ctx := context.Background()

	repo := git.NewLocalRepository()
	defer repo.Close()

	err := repo.Open(ctx, ".")
	require.NoError(t, err)

	branches, err := git.ListBranches(ctx, repo)
	require.NoError(t, err)

	activeBranches := git.FilterActiveBranches(branches)
	assert.NotEmpty(t, activeBranches)
}
