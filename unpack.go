package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

func getFilename(assetUrl string, blueprint Blueprint) string {
	for _, asset := range blueprint.Props {
		if asset == nil {
			continue
		}
		a, ok := asset.(map[string]any)
		if !ok {
			continue
		}

		if a["url"] == assetUrl {
			return a["name"].(string)
		}
	}
	return ""
}

func unpackHyp(filename string) {
	blob, err := os.ReadFile(filename)
	if err != nil {
		panic(err)
	}

	blueprint, assets, err := ImportApp(blob)
	if err != nil {
		panic(err)
	}

	fmt.Printf("Total assets %d\n", len(assets))
	os.Mkdir(blueprint.Name, 0777)

	for _, asset := range assets {
		fmt.Printf("%s - %s - %d\n", asset.URL, asset.Type, asset.Size)
		filename := strings.Split(asset.URL, "//")[1]
		ext := strings.Split(asset.URL, ".")[1]

		if asset.Type == "script" {
			filename = "script.js"
		} else if asset.URL == blueprint.Model {
			filename = fmt.Sprintf("model.%s", ext)
		} else if blueprint.Image != nil && asset.URL == blueprint.Image.URL {
			filename = blueprint.Image.Name
		} else {
			loc_file := getFilename(asset.URL, *blueprint)
			if loc_file != "" {
				filename = loc_file
			}
		}

		os.WriteFile(fmt.Sprintf("./%s/%s", blueprint.Name, filename), asset.FileData, 0777)
	}

	hdr := HypeHeader{
		Blueprint: blueprint,
		Assets:    assets,
	}

	json_data, err := json.MarshalIndent(hdr, "", "    ")
	os.WriteFile(fmt.Sprintf("./%s/header.json", blueprint.Name), json_data, 0777)
}
