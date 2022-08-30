package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"ferlab/terracd/fs"
)

func handleErr(err error) {
	if err != nil {
		panic(err.Error())
	}
}

func cleanup(workDir string, stateDir string) {
	stateSrc := path.Join(workDir, "terraform.tfstate")
	stateDest := path.Join(stateDir, "terraform.tfstate")
	stateBackSrc := path.Join(workDir, "terraform.tfstate.backup")
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

	removalErr := os.RemoveAll(workDir)
	handleErr(removalErr)
}

func main() {
	wd, wdErr := os.Getwd()
	handleErr(wdErr)

	reposDir := path.Join(wd, "repos")
	stateDir := path.Join(wd, "state")
	workDir := path.Join(wd, "work")

	workDirExists, workDirExistsErr := fs.PathExists(workDir)
	handleErr(workDirExistsErr)
	if workDirExists {
		fmt.Println("Warning: Working directory found from prior iteration. Will clean it up.")
		cleanup(workDir, stateDir)
	}
	fs.AssurePrivateDir(reposDir)
	fs.AssurePrivateDir(stateDir)
	fs.AssurePrivateDir(workDir)
	defer cleanup(workDir, stateDir)

	config, configErr := getConfig()
	handleErr(configErr)

	syncErr := syncConfigRepos(reposDir, config)
	handleErr(syncErr)

	mergeErr := fs.MergeDirs(workDir, append(getSourcePaths(reposDir, config), stateDir))
	handleErr(mergeErr)

    switch config.Command {
    case "wait":
		waitTime := config.Timeouts.Wait
		if int64(waitTime) == int64(0) {
			waitTime, _ = time.ParseDuration("1h")
		}
		time.Sleep(waitTime)
    case "plan":
		planErr := terraformPlan(workDir, config)
		handleErr(planErr)
    case "apply":
		applyErr := terraformApply(workDir, config)
		handleErr(applyErr)
    case "migrate_backend":
		migrateErr := terraformMigrateBackend(workDir, config)
		handleErr(migrateErr)
    }
}