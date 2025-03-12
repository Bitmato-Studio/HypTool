package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type MetaData struct {
	ID      string `json:"id"`
	Name    string `json:"name"`
	Version int    `json:"version"` // this is here because hyperfy requires it
	Author  string `json:"author"`
	URL     string `json:"url"`
	Desc    string `json:"desc"`
	Model   string `json:"model"`

	Preload bool `json:"preload"`
	Public  bool `json:"public"`
	Unique  bool `json:"unique"`
}

type Config struct {
	Data       MetaData `json:"data"`
	AppVersion string   `json:"app_version"`
	ScriptPath string   `json:"script_path"`
	AssetsPath string   `json:"assets_path"`
	PropsPath  string   `json:"props_path"`
}

func LoadConfig(path string) *Config {
	blob, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	var conf Config
	if err := json.Unmarshal(blob, &conf); err != nil {
		panic(err)
	}

	return &conf
}

func LoadConfigMHA(path string) *[]Config {
	blob, err := os.ReadFile(path)

	if err != nil {
		panic(err)
	}

	var confs []Config
	if err := json.Unmarshal(blob, &confs); err != nil {
		panic(err)
	}

	return &confs
}

func SaveMHAConfig(path string, configs *[]Config) error {
	file, err := os.Create(path)

	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(configs, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}
	return nil
}

func SaveConfig(path string, config *Config) error {
	file, err := os.Create(path)

	if err != nil {
		fmt.Println("Error creating file:", err)
		return err
	}
	defer file.Close()

	jsonData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		fmt.Println("Error marshalling JSON:", err)
		return err
	}

	_, err = file.Write(jsonData)
	if err != nil {
		fmt.Println("Error writing file:", err)
		return err
	}
	return nil
}
