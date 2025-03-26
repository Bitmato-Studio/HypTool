package main

import (
	"fmt"
	"os"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
)

func CloneRepo(path string, url string, branch string) {
	branchRef := plumbing.NewBranchReferenceName(branch)
	fmt.Println("🔄 Cloning or updating repository into:", path)

	// Try opening the repository. If it doesn't exist, clone it.
	repo, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		repo, err = git.PlainClone(path, false, &git.CloneOptions{
			URL:           url,
			ReferenceName: branchRef,
			SingleBranch:  true,
			Progress:      os.Stdout,
		})
		if err != nil {
			fmt.Println("Error cloning repository:", err)
			panic(err)
		}
		fmt.Println("Repository cloned successfully into", path)
		return
	} else if err != nil {
		fmt.Println("Error opening repository:", err)
		panic(err)
	}

	// If the repo exists, ensure the remote "origin" is set up.
	remote, err := repo.Remote("origin")
	if err != nil {
		// Remote not found: create it.
		_, err = repo.CreateRemote(&config.RemoteConfig{
			Name: "origin",
			URLs: []string{url},
		})
		if err != nil {
			fmt.Println("Error creating remote:", err)
			panic(err)
		}
	} else {
		// Optionally, update remote URL if necessary.
		_ = remote
	}

	// Get the worktree to work with files.
	w, err := repo.Worktree()
	if err != nil {
		fmt.Println("Error getting worktree:", err)
		panic(err)
	}

	// Attempt to check out the desired branch.
	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchRef,
	})
	if err != nil {
		// If checkout fails (e.g., branch doesn't exist locally), create a new branch.
		err = w.Checkout(&git.CheckoutOptions{
			Branch: branchRef,
			Create: true,
		})
		if err != nil {
			fmt.Println("Error checking out branch:", err)
			panic(err)
		}
	}

	// Pull the latest changes from the branch.
	err = w.Pull(&git.PullOptions{
		RemoteName:    "origin",
		ReferenceName: branchRef,
		SingleBranch:  true,
		Progress:      os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Println("Error pulling repository:", err)
		panic(err)
	}

	fmt.Println("Repository updated successfully on branch", branch)
}
