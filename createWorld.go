package main

import "os"

func createWorld(gitRepo string, branch string) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	CloneRepo(dir, gitRepo, branch)
}
