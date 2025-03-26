package main

import (
	"flag"
	"log"
	"os"
	"strings"
)

const APPROLLUP_FILENAME = "approllup.json"
const APPROLLUP_MHA_NAME = "approllup.mha.json"

func main() {
	isInitApp := flag.Bool("init", false, "Create a new app project")
	isInitMultiApp := flag.Bool("initmha", false, "Create a new app with multiple apps")
	isBuildApp := flag.Bool("build", false, "Build a App-Rollup")

	isAddProp := flag.Bool("prop", false, "Run build a prop")
	isAddApp := flag.Bool("addapp", false, "Add a new app to multi-hyp app")

	// isWorldAction := flag.Bool("world", false, "Toggle if we're interacting with a world")
	// isCreateWorld := flag.Bool("new_world", false, "If we are creating a new local world")
	// isCloneWorld := flag.Bool("clone_world", false, "If we are creating a new local clone")
	// hyperfyBranch := flag.String("hyperfy_branch", "dev", "What branch do you want to use")
	// hyperfyRepo := flag.String("hyperfy_branch", "github.com/hyperfy-xyz/hyperfy", "What branch do you want to use")
	// cloneUrl := flag.String("clone_uri", "localhost:3000", "What uri you want to clone from")

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
		} else {
			buildMHAProject(*defaultUnique, *buildHypJson, *scriptBuild)
		}
	}

	if *isAddProp {
	}

	// if *isCreateWorld {
	// 	createWorld(*hyperfyRepo, *hyperfyBranch)
	// }

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
