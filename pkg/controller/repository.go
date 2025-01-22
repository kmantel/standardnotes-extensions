package controller

import (
	"errors"
	"fmt"
	"log"
	"time"

	"github.com/go-git/go-git/v5"
)

func RepoUpdate(dir, url string, pull_opts *git.PullOptions) (*git.Repository, error) {
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
	err = worktree.Pull(pull_opts)
	if err == nil {
		log.Printf("found new update: %q\n", url)
		return repo, nil
	}
	if !errors.Is(err, git.NoErrAlreadyUpToDate) {
		return nil, fmt.Errorf("pull worktree: %w", err)
	}
	return repo, nil
}

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
