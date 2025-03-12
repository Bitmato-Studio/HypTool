package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

const APPROLLUP_FILENAME = "approllup.json"
const APPROLLUP_MHA_NAME = "approllup.mha.json"

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
	for _, asset := range assets {
		fmt.Printf("%s - %s - %d\n", asset.URL, asset.Type, asset.Size)

		os.Mkdir("./Unpacked", 0777)

		os.WriteFile(fmt.Sprintf("./Unpacked/%s", strings.Split(asset.URL, "//")[1]), asset.FileData, 0777)
	}

	hdr := HypeHeader{
		Blueprint: blueprint,
		Assets:    assets,
	}

	json_data, err := json.MarshalIndent(hdr, "", "    ")
	os.WriteFile("./Unpacked/header.json", json_data, 0777)
}

func main() {
	isInitApp := flag.Bool("init", false, "Create a new app project")
	isInitMultiApp := flag.Bool("initmha", false, "Create a new app with multiple apps")
	isBuildApp := flag.Bool("build", false, "Build a App-Rollup")
	isAddProp := flag.Bool("prop", false, "Run build a prop")
	isAddApp := flag.Bool("addapp", false, "Add a new app to multi-hyp app")

	scriptBuild := flag.Bool("nsb", false, "Turns off npx rollup -c command")
	isUnpack := flag.Bool("unpack", false, "Run unpack")
	filepath := flag.String("file", "", "The file to manipulate")

	buildHypJson := flag.Bool("bjson", false, "Build App json for debugging")
	defaultUnique := flag.Bool("dunique", true, "Disable unique by default")

	isSetVersion := flag.Bool("setversion", false, "Run to change the version of your app")

	flag.Parse()
	if *isInitApp {
		runCreateApp(false)
	}

	if *isInitMultiApp {
		runCreateApp(true)
	}

	if *isAddApp {
		createMHASub()
	}

	if *isBuildApp {
		dir, err := os.Getwd()
		if err != nil {
			panic(err)
		}

		files, err := os.ReadDir(dir)
		if err != nil {
			panic(err)
		}

		isMHA := false

		for _, f := range files {
			if strings.Contains(f.Name(), "mha") {
				isMHA = true
				break
			}
		}

		if !isMHA {
			buildAppProject(*defaultUnique, *buildHypJson, *scriptBuild, nil)
			return
		}
		buildMHAProject(*defaultUnique, *buildHypJson, *scriptBuild)
	}

	if *isAddProp {
	}

	if *isSetVersion {
		args := flag.Args()
		if len(args) < 1 {
			log.Fatal("Error: you must specify a version as a positional argument.")
		}
	}

	if *isUnpack {
		unpackHyp(*filepath)
	}

}
