package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
)

type Prop map[string]any
type Props []Prop

func loadProps(path string) Props {
	var result Props

	blob, err := os.ReadFile(path)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(blob, &result)
	if err != nil {
		panic(err)
	}

	return result
}

func buildBlueprintProps(props Props, header *HypeHeader) {
	bp := header.Blueprint

	for _, prop := range props {
		key := prop["key"].(string)
		if _, exists := prop["initial"]; !exists {
			bp.Props[key] = "" // Needa check prop["type"]
			continue
		}

		bp.Props[key] = prop["initial"].(string)
	}
}

func addScript(header *HypeHeader, config *Config) {
	fmt.Printf("Script were built for %s\n", config.Data.Name)

	fmt.Printf("Adding script %s to %s\n", config.ScriptPath, config.Data.Name)
	scriptBlob, err := os.ReadFile(config.ScriptPath)
	if err != nil {
		panic(err)
	}

	script_asset, err := AddAssetToGroup(&header.Assets, scriptBlob, "script")
	if err != nil {
		panic(err)
	}

	header.Blueprint.Script = script_asset.URL

	fmt.Printf("Script added to project %s\n", config.Data.Name)
}

func addModel(header *HypeHeader, config *Config) {
	fmt.Printf("Building Model %s\n", config.Data.Model)

	model_blob, err := os.ReadFile(config.Data.Model)
	if err != nil {
		panic(err)
	}

	model_type := "model"
	if strings.HasSuffix(config.Data.Model, ".vrm") {
		model_type = "avatar"
	}

	model_asset, err := AddAssetToGroup(&header.Assets, model_blob, model_type)
	if err != nil {
		panic(err)
	}
	header.Blueprint.Model = model_asset.URL
}

func buildPropFile(header *HypeHeader, prop map[string]any) {

	if _, initialExists := prop["initial"]; !initialExists {
		fmt.Printf("%s has no \"initial\"\n", prop["key"].(string))
		// no file to prebuild
		return
	}

	if _, kindExists := prop["kind"]; !kindExists {
		panic(fmt.Errorf("Key \"kind\" was not found on prop for %s", prop["key"].(string)))
	}

	data := map[string]string{
		"type": prop["type"].(string),
		"name": prop["initial"].(string),
		"url":  "",
	}

	fileBlob, err := os.ReadFile(prop["initial"].(string))
	if err != nil {
		panic(err)
	}

	asset, err := AddAssetToGroup(&header.Assets, fileBlob, prop["kind"].(string))
	if err != nil {
		panic(err)
	}

	data["url"] = asset.URL
	header.Blueprint.Props[prop["key"].(string)] = data
}

func validateProp(prop map[string]any) bool {
	if _, keyExists := prop["key"]; !keyExists {
		fmt.Println("Required key 'key' is missing!")
		return false
	}

	if _, typeExists := prop["type"]; !typeExists {
		fmt.Println("Required key 'type' is missing!")
		return false
	}
	return true
}

func buildProps(header *HypeHeader, config *Config) {
	props_blob, err := os.ReadFile(config.PropsPath)
	if err != nil {
		panic(err)
	}

	var data []map[string]any
	err = json.Unmarshal(props_blob, &data)
	if err != nil {
		panic(err)
	}

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, prop := range data {
		prop := prop
		wg.Add(1)

		go func(prop map[string]any) {
			defer wg.Done()

			if !validateProp(prop) {
				panic(fmt.Errorf("Prop does not meet standard see docs"))
			}
			fmt.Printf("Building %s prop of type %s\n", prop["key"], prop["type"])

			if prop["type"] == "file" {
				buildPropFile(header, prop)
				return
			}

			if initial, initialExists := prop["initial"]; initialExists {
				mutex.Lock()
				header.Blueprint.Props[prop["key"].(string)] = initial
				mutex.Unlock()
			}
		}(prop)
	}

	wg.Wait()
}

func buildAppProject(uniqueDefault bool, buildHypJson bool, noScriptBuild bool) {
	dir, err := os.Getwd()

	if err != nil {
		panic(err)
	}

	rollup_config := filepath.Join(dir, APPROLLUP_FILENAME)

	config := LoadConfig(rollup_config)

	// props_path := filepath.Join(dir, config.PropsPath)
	// props := loadProps(props_path)

	fmt.Printf("Building app %s by %s (v%d)\n", config.Data.Name, config.Data.Author, config.Data.Version)

	var newHeader HypeHeader

	id := config.Data.ID
	if id == "" {
		id = uuid()
	}

	newHeader.Blueprint = &Blueprint{
		ID:      id,
		Name:    config.Data.Name,
		Author:  config.Data.Author,
		Version: config.Data.Version,
		Desc:    config.Data.Desc,
		URL:     config.Data.URL,

		Model:  "",
		Script: "",
		Props:  map[string]any{},

		Unique:  config.Data.Unique || uniqueDefault,
		Locked:  false,
		Frozen:  false,
		Preload: config.Data.Preload,
		Public:  config.Data.Public,
	}

	fmt.Printf("Building %s's scripts\n", config.Data.Name)

	/* Build the scripts */
	if !noScriptBuild {
		cmd := exec.Command("npx", "rollup", "-c")
		cmd.Stdout = os.Stdout // Pipe output to terminal
		cmd.Stderr = os.Stderr // Pipe errors to terminal

		err = cmd.Run()
		if err != nil {
			panic(err)
		}
	}

	// -- ADD SCRIPT TO HYP --
	addScript(&newHeader, config)

	// -- ADD MODEL TO HYP --
	addModel(&newHeader, config)

	// -- ADD PROPS TO HYP --
	buildProps(&newHeader, config)

	// -- DONE BUILDING -- //

	// Bundle the hype
	fmt.Printf("We have %d assets for %s\n", len(newHeader.Assets), newHeader.Blueprint.Name)
	hyp_data, filename, err := ExportApp(newHeader.Blueprint, newHeader.Assets)
	if err != nil {
		panic(err)
	}

	// Save stuff to hyp
	file, err := os.Create(filename)
	if err != nil {
		panic(err)
	}
	defer file.Close()

	file.Write(hyp_data)

	// build hyp json
	if buildHypJson {
		file, err = os.Create(filename + ".json")
		if err != nil {
			panic(err)
		}
		defer file.Close()

		blob, err := json.MarshalIndent(newHeader, "", "    ")
		if err != nil {
			panic(err)
		}
		file.Write(blob)
	}
}
