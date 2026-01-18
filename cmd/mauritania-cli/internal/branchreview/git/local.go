package git

import (
	"context"
	"fmt"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

// LocalRepository implements Repository interface for local Git repositories
type LocalRepository struct {
	repo *git.Repository
}

// NewLocalRepository creates a new LocalRepository instance
func NewLocalRepository() *LocalRepository {
	return &LocalRepository{}
}

// Open opens a Git repository at the given path
func (lr *LocalRepository) Open(ctx context.Context, path string) error {
	repo, err := git.PlainOpen(path)
	if err != nil {
		return fmt.Errorf("failed to open repository at %s: %w", path, err)
	}
	lr.repo = repo
	return nil
}

// GetBranches returns all branches in the repository
func (lr *LocalRepository) GetBranches(ctx context.Context) ([]BranchInfo, error) {
	if lr.repo == nil {
		return nil, fmt.Errorf("repository not opened")
	}

	branches := []BranchInfo{}

	// Get local branches
	localBranches, err := lr.repo.Branches()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	err = localBranches.ForEach(func(ref *plumbing.Reference) error {
		branchName := ref.Name().Short()

		commit, err := lr.repo.CommitObject(ref.Hash())
		if err != nil {
			return fmt.Errorf("failed to get commit for branch %s: %w", branchName, err)
		}

		commitInfo := &CommitInfo{
			Hash:        commit.Hash.String(),
			AuthorName:  commit.Author.Name,
			AuthorEmail: commit.Author.Email,
			Message:     commit.Message,
			Date:        commit.Author,
		}

		branchInfo := BranchInfo{
			Name:       branchName,
			Hash:       ref.Hash().String(),
			IsRemote:   false,
			LastCommit: commitInfo,
		}

		branches = append(branches, branchInfo)
		return nil
	})

	if err != nil {
		return nil, err
	}

	// Get remote branches
	remotes, err := lr.repo.Remotes()
	if err != nil {
		return nil, fmt.Errorf("failed to get remotes: %w", err)
	}

	for _, remote := range remotes {
		remoteRefs, err := remote.List(&git.ListOptions{})
		if err != nil {
			continue // Skip remotes that can't be listed
		}

		for _, ref := range remoteRefs {
			if ref.Name().IsBranch() {
				branchName := ref.Name().Short()

				commit, err := lr.repo.CommitObject(ref.Hash())
				if err != nil {
					continue
				}

				commitInfo := &CommitInfo{
					Hash:        commit.Hash.String(),
					AuthorName:  commit.Author.Name,
					AuthorEmail: commit.Author.Email,
					Message:     commit.Message,
					Date:        commit.Author,
				}

				branchInfo := BranchInfo{
					Name:       branchName,
					Hash:       ref.Hash().String(),
					IsRemote:   true,
					LastCommit: commitInfo,
				}

				branches = append(branches, branchInfo)
			}
		}
	}

	return branches, nil
}

// GetCommits returns commits for a specific branch
func (lr *LocalRepository) GetCommits(ctx context.Context, branchName string, limit int) ([]CommitInfo, error) {
	if lr.repo == nil {
		return nil, fmt.Errorf("repository not opened")
	}

	branchRef, err := lr.repo.Branch(branchName)
	if err != nil {
		return nil, fmt.Errorf("failed to get branch %s: %w", branchName, err)
	}

	hash := plumbing.NewHash(branchRef.Hash.String())
	if hash.IsZero() {
		return nil, fmt.Errorf("invalid branch hash")
	}

	commitIter, err := lr.repo.Log(&git.LogOptions{
		From:  hash,
		Order: git.LogOrderCommitterTime,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get commit log: %w", err)
	}
	defer commitIter.Close()

	commits := []CommitInfo{}
	count := 0

	err = commitIter.ForEach(func(commit *object.Commit) error {
		if limit > 0 && count >= limit {
			return fmt.Errorf("limit reached") // Break iteration
		}

		parents := []string{}
		for _, parent := range commit.ParentHashes {
			parents = append(parents, parent.String())
		}

		// Get changed files and stats
		stats, err := commit.Stats()
		if err != nil {
			return fmt.Errorf("failed to get commit stats: %w", err)
		}

		changedFiles := []string{}
		insertions := 0
		deletions := 0

		for _, stat := range stats {
			changedFiles = append(changedFiles, stat.Name)
			insertions += stat.Addition
			deletions += stat.Deletion
		}

		commitInfo := CommitInfo{
			Hash:         commit.Hash.String(),
			AuthorName:   commit.Author.Name,
			AuthorEmail:  commit.Author.Email,
			Message:      commit.Message,
			Date:         commit.Author,
			Parents:      parents,
			ChangedFiles: changedFiles,
			Insertions:   insertions,
			Deletions:    deletions,
		}

		commits = append(commits, commitInfo)
		count++
		return nil
	})

	if err != nil && err.Error() != "limit reached" {
		return nil, err
	}

	return commits, nil
}

// GetCommit returns a specific commit by hash
func (lr *LocalRepository) GetCommit(ctx context.Context, hash string) (*CommitInfo, error) {
	if lr.repo == nil {
		return nil, fmt.Errorf("repository not opened")
	}

	commit, err := lr.repo.CommitObject(plumbing.NewHash(hash))
	if err != nil {
		return nil, fmt.Errorf("failed to get commit %s: %w", hash, err)
	}

	parents := []string{}
	for _, parent := range commit.ParentHashes {
		parents = append(parents, parent.String())
	}

	stats, err := commit.Stats()
	if err != nil {
		return nil, fmt.Errorf("failed to get commit stats: %w", err)
	}

	changedFiles := []string{}
	insertions := 0
	deletions := 0

	for _, stat := range stats {
		changedFiles = append(changedFiles, stat.Name)
		insertions += stat.Addition
		deletions += stat.Deletion
	}

	return &CommitInfo{
		Hash:         commit.Hash.String(),
		AuthorName:   commit.Author.Name,
		AuthorEmail:  commit.Author.Email,
		Message:      commit.Message,
		Date:         commit.Author,
		Parents:      parents,
		ChangedFiles: changedFiles,
		Insertions:   insertions,
		Deletions:    deletions,
	}, nil
}

// Close closes the repository connection
func (lr *LocalRepository) Close() error {
	// Local repository doesn't need explicit closing
	lr.repo = nil
	return nil
}
