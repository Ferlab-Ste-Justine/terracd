package main

import (
	"os"
	"path"

	"ferlab/terracd/fs"
)

func handleErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func cleanup(applyDir string, stateDir string) {
	stateSrc := path.Join(applyDir, "terraform.tfstate")
	stateDest := path.Join(stateDir, "terraform.tfstate")
	stateBackSrc := path.Join(applyDir, "terraform.tfstate.backup")
	stateBackDest := path.Join(stateDir, "terraform.tfstate.backup")

	stateSrcExists, stateSrcErr := fs.PathExists(stateSrc)
	handleErr(stateSrcErr)

	stateBackSrcExists, stateBackSrcErr := fs.PathExists(stateBackSrc)
	handleErr(stateBackSrcErr)

	if stateBackSrcExists {
		fs.CopyPrivateFile(stateBackSrc, stateBackDest)
	}

	if stateSrcExists {
		fs.CopyPrivateFile(stateSrc, stateDest)
	}

	removalErr := os.RemoveAll(applyDir)
	handleErr(removalErr)
}

func main() {
	wd, wdErr := os.Getwd()
	handleErr(wdErr)

	reposDir := path.Join(wd, "repos")
	stateDir := path.Join(wd, "state")
	applyDir := path.Join(wd, "apply")

	applyExists, applyExistsErr := fs.PathExists(applyDir)
	handleErr(applyExistsErr)
	if applyExists {
		panic("Error: Apply directory already exists.")
	}
	fs.AssurePrivateDir(reposDir)
	fs.AssurePrivateDir(stateDir)
	fs.AssurePrivateDir(applyDir)
	defer cleanup(applyDir, stateDir)

	config, configErr := getConfig()
	handleErr(configErr)

	syncErr := syncConfigRepos(reposDir, config)
	handleErr(syncErr)

	mergeErr := fs.MergeDirs(applyDir, append(getSourcePaths(reposDir, config), stateDir))
	handleErr(mergeErr)

	applyErr := terraformApply(applyDir, config)
	handleErr(applyErr)
}