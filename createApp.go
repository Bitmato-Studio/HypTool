package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/manifoldco/promptui"
)

const TEMPLATE_URL = "https://github.com/Bitmato-Studio/App-Rollup"

func CloneRepo(path string) {
	fmt.Println("🔄 Cloning repository into:", path)

	// Check if this directory is already a Git repository
	repo, err := git.PlainOpen(path)
	if err == git.ErrRepositoryNotExists {
		// Initialize a new Git repo
		repo, err = git.PlainInit(path, false)
		if err != nil {
			fmt.Println("Error initializing repository:", err)
			panic(err)
		}
	}

	_, err = repo.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{TEMPLATE_URL},
	})
	if err != nil {
		fmt.Println("Remote already exists or error adding remote:", err)
		panic(err)
	}

	// Pull the latest changes
	w, err := repo.Worktree()
	if err != nil {
		fmt.Println("Error getting worktree:", err)
		panic(err)
	}

	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})
	if err != nil && err != git.NoErrAlreadyUpToDate {
		fmt.Println("Error pulling repository:", err)
		panic(err)
	}

	fmt.Println("Repository cloned successfully into", path)
}

func generateConfig() *Config {
	metadata := MetaData{
		ID:      uuid(),
		Name:    promptInput("App Name", "my-app-project"),
		Version: 1,
		Author:  promptInput("Author", "Your Name"),
		URL:     promptInput("App URL", "https://example.com"),
		Desc:    promptInput("Description", "A new app project"),
		Model:   "./assets/model.glb",

		Unique:  true,
		Preload: false,
		Public:  false,
	}

	config := Config{
		Data:       metadata,
		AppVersion: promptInput("Version", "v1.0.0"),
		ScriptPath: "./dist/main.bundle.js",
		AssetsPath: "./assets",
		PropsPath:  "./props/props.json",
	}

	return &config
}

func createMHASub() {
	root, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	new_config := generateConfig()
	mha_path := filepath.Join(root, APPROLLUP_MHA_NAME)
	configs := LoadConfigMHA(mha_path)
	*configs = append(*configs, *new_config)

	app_dir := filepath.Join(root, new_config.Data.Name)
	os.MkdirAll(app_dir, 0777)

	CloneRepo(app_dir)

	err = SaveMHAConfig(mha_path, configs)
	if err != nil {
		panic(err)
	}
}

func runCreateApp(isMHA bool) {
	dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}

	root_name := ""
	rollup_name := APPROLLUP_FILENAME

	if isMHA {
		root_name = promptInput("Root Name for MHA", "root")
		rollup_name = APPROLLUP_MHA_NAME
	}

	config := generateConfig()

	app_dir := filepath.Join(dir, config.Data.Name)
	config_path := filepath.Join(app_dir, rollup_name)
	if isMHA {
		app_dir = filepath.Join(dir, root_name, config.Data.Name)
		config_path = filepath.Join(dir, root_name, rollup_name)
	}

	os.MkdirAll(app_dir, 0777)

	CloneRepo(app_dir)

	if !isMHA {
		err = SaveConfig(config_path, config)
	} else {
		var confs []Config
		confs = append(confs, *config)
		err = SaveMHAConfig(config_path, &confs)
	}

	if err != nil {
		panic(err)
	}

	fmt.Println("✅ Project initialized! Configuration saved to config.json")
}

func promptInput[T any](label string, defaultValue T) T {
	prompt := promptui.Prompt{
		Label:   label,
		Default: fmt.Sprintf("%v", defaultValue),
	}

	result, err := prompt.Run()
	if err != nil {
		fmt.Println("Prompt failed:", err)
		os.Exit(1)
	}

	// Convert result to the appropriate type
	var finalValue T
	switch any(defaultValue).(type) {
	case int:
		num, err := strconv.Atoi(result)
		if err != nil {
			fmt.Println("Invalid number input, using default value:", defaultValue)
			return defaultValue
		}
		finalValue = any(num).(T)
	default:
		finalValue = any(result).(T)
	}

	return finalValue
}
