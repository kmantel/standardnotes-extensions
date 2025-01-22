package controller

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/config"
)


type RepoUpdateOptions struct {
    Revision string
}

func RepoUpdate(dir, url string, pullOpts *git.PullOptions, opts *RepoUpdateOptions) (*git.Repository, error) {
	repo, err := git.PlainClone(dir, false, &git.CloneOptions{
		URL: url,
	})
	if err == nil {
		log.Printf("found new repo: %q\n", url)
		return repo, nil
	}
	if !errors.Is(err, git.ErrRepositoryAlreadyExists) {
		return nil, fmt.Errorf("plain clone: %w", err)
	}
	repo, err = git.PlainOpen(dir)
	if err != nil {
		return nil, fmt.Errorf("plain open: %w", err)
	}
	worktree, err := repo.Worktree()
	if err != nil {
		return nil, fmt.Errorf("open worktree: %w", err)
	}

	if opts.Revision != "" {
		revision := opts.Revision
		// err := RepoCheckout(worktree, opts.Revision)
		// https://github.com/go-git/go-git/issues/279
		localBranchLookup, err := repo.Branch(revision)
		localReferenceName := plumbing.NewBranchReferenceName(revision)
		if err != nil {

			remoteReferenceName := plumbing.NewRemoteReferenceName("origin", revision)
			err := repo.CreateBranch(&config.Branch{Name: revision, Remote: "origin", Merge: localReferenceName })
			if err != nil {
				return nil, fmt.Errorf("repo create branch: %s %w\n", url, err)
			}
			newReference := plumbing.NewSymbolicReference(localReferenceName , remoteReferenceName)
			err = repo.Storer.SetReference(newReference)
		} else {
			localReferenceName = localBranchLookup.Merge
		}
		err = worktree.Checkout(&git.CheckoutOptions{
			Branch: plumbing.ReferenceName(localReferenceName.String()),
		})
		if err != nil {
			return nil, fmt.Errorf("repo checkout: %w %s", err, revision)
		}
		pullOpts.ReferenceName = localReferenceName
	}

	err = worktree.Pull(pullOpts)
	if err == nil {
		log.Printf("found new update: %q\n", url)
		return repo, nil
	}
	if !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, fmt.Errorf("pull worktree: %w", err)
	}
	return repo, nil
}

// func RepoCheckout(repo *git.Repository, revision string) (error) {

// }

func RepoGetHEAD(repo *git.Repository) (string, error) {
	head, err := repo.Head()
	if err != nil {
		return "", fmt.Errorf("get head: %w", err)
	}
	return head.Hash().String()[:8], nil
}

func RepoGetLatestStamp(repo *git.Repository) (time.Time, error) {
	iter, err := repo.Log(&git.LogOptions{})
	if err != nil {
		return time.Time{}, fmt.Errorf("log repo: %w", err)
	}
	latest, err := iter.Next()
	if err != nil {
		return time.Time{}, fmt.Errorf("get latest: %w", err)
	}
	return latest.Committer.When, nil
}
