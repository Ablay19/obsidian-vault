package models

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewBranch(t *testing.T) {
	branch := NewBranch("feature/test", "https://github.com/user/repo")

	assert.Equal(t, "feature/test", branch.Name)
	assert.Equal(t, "https://github.com/user/repo", branch.RepositoryURL)
	assert.Equal(t, BranchStatusActive, branch.Status)
	assert.Equal(t, 0, branch.AheadCount)
	assert.Equal(t, 0, branch.BehindCount)
	assert.NotZero(t, branch.CreatedAt)
	assert.NotZero(t, branch.UpdatedAt)
}

func TestBranchValidate(t *testing.T) {
	tests := []struct {
		name    string
		branch  *Branch
		wantErr bool
	}{
		{
			name: "valid branch",
			branch: &Branch{
				Name:          "main",
				RepositoryURL: "https://github.com/user/repo",
			},
			wantErr: false,
		},
		{
			name: "empty name",
			branch: &Branch{
				RepositoryURL: "https://github.com/user/repo",
			},
			wantErr: true,
		},
		{
			name: "empty repo URL",
			branch: &Branch{
				Name: "main",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.branch.Validate()
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestBranchStatusTransitions(t *testing.T) {
	// Test that status can be changed
	branch := NewBranch("test", "repo")
	branch.Status = BranchStatusStale
	assert.Equal(t, BranchStatusStale, branch.Status)
}
