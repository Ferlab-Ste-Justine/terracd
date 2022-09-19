package main

import (
	"fmt"
	"os"
	"path"
	"time"

	"ferlab/terracd/fs"
)

func cleanup(workDir string, stateDir string) error {
	stateSrc := path.Join(workDir, "terraform.tfstate")
	stateDest := path.Join(stateDir, "terraform.tfstate")
	stateBackSrc := path.Join(workDir, "terraform.tfstate.backup")
	stateBackDest := path.Join(stateDir, "terraform.tfstate.backup")

	stateSrcExists, stateSrcErr := fs.PathExists(stateSrc)
	if stateSrcErr != nil {
		return stateSrcErr
	}

	stateBackSrcExists, stateBackSrcErr := fs.PathExists(stateBackSrc)
	if stateBackSrcErr != nil {
		return stateBackSrcErr
	}

	if stateBackSrcExists {
		copyErr := fs.CopyPrivateFile(stateBackSrc, stateBackDest)
		if copyErr != nil {
			return copyErr
		}
	}

	if stateSrcExists {
		copyErr := fs.CopyPrivateFile(stateSrc, stateDest)
		if copyErr != nil {
			return copyErr
		}
	}

	removalErr := os.RemoveAll(workDir)
	if removalErr != nil {
		return removalErr
	}

	return nil
}

func Exec() error {
	wd, wdErr := os.Getwd()
	if wdErr != nil {
		return wdErr
	}

	reposDir := path.Join(wd, "repos")
	stateDir := path.Join(wd, "state")
	workDir := path.Join(wd, "work")

	workDirExists, workDirExistsErr := fs.PathExists(workDir)
	if workDirExistsErr != nil {
		return workDirExistsErr
	}
	if workDirExists {
		fmt.Println("Warning: Working directory found from prior iteration. Will clean it up.")
		cleanupErr := cleanup(workDir, stateDir)
		if cleanupErr != nil {
			return cleanupErr
		}
	}

	assureErr := fs.AssurePrivateDir(reposDir)
	if assureErr != nil {
		return assureErr
	}

	assureErr = fs.AssurePrivateDir(stateDir)
	if assureErr != nil {
		return assureErr
	}

	assureErr = fs.AssurePrivateDir(workDir)
	if assureErr != nil {
		return assureErr
	}

	defer func() {
		cleanupErr := cleanup(workDir, stateDir)
		if cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup working directory at the end of execution: %s.\n", cleanupErr.Error())
		}
	}()

	config, configErr := getConfig()
	if configErr != nil {
		return configErr
	}

	syncErr := syncConfigRepos(reposDir, config)
	if syncErr != nil {
		return syncErr
	}

	mergeErr := fs.MergeDirs(workDir, append(getSourcePaths(reposDir, config), stateDir))
	if mergeErr != nil {
		return mergeErr
	}

    switch config.Command {
    case "wait":
		waitTime := config.Timeouts.Wait
		if int64(waitTime) == int64(0) {
			waitTime, _ = time.ParseDuration("1h")
		}
		time.Sleep(waitTime)
    case "plan":
		planErr := terraformPlan(workDir, config)
		if planErr != nil {
			return planErr
		}
    case "apply":
		applyErr := terraformApply(workDir, config)
		if applyErr != nil {
			return applyErr
		}
    case "migrate_backend":
		migrateErr := terraformMigrateBackend(workDir, config)
		if migrateErr != nil {
			return migrateErr
		}
    }

	return nil
}

func main() {
	err := Exec()
	if err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}