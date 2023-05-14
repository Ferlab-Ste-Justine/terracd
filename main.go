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

func getSourcePaths(repoDir string, c Config) []string {
	paths := []string{}
	for _, source := range c.Sources {
		if source.Repo.Url != "" {
			dir := getRepoDir(source.Repo.Url, source.Repo.Ref)
			dir = path.Join(repoDir, dir)
			if source.Repo.Path != "" {
				dir = path.Join(dir, source.Repo.Path)
			}
			paths = append(paths, dir)
		} else if source.Dir != "" {
			paths = append(paths, source.Dir)
		}
	}

	return paths
}

func Exec(config Config) error {
	workDirExists, workDirExistsErr := fs.PathExists(config.WorkingDirectory)
	if workDirExistsErr != nil {
		return workDirExistsErr
	}
	if !workDirExists {
		assureErr := fs.AssurePrivateDir(config.WorkingDirectory)
		if assureErr != nil {
			return assureErr
		}
	}
	chdirErr := os.Chdir(config.WorkingDirectory)
	if chdirErr != nil {
		return chdirErr
	}

	reposDir := path.Join(config.WorkingDirectory, "repos")
	backendDir := path.Join(config.WorkingDirectory, "backend")
	stateDir := path.Join(config.WorkingDirectory, "state")
	workDir := path.Join(config.WorkingDirectory, "work")

	workDirExists, workDirExistsErr = fs.PathExists(workDir)
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

	assureErr = fs.AssurePrivateDir(backendDir)
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

	syncErr := syncConfigRepos(reposDir, config)
	if syncErr != nil {
		return syncErr
	}

	backendGenErr := GenerateBackendFiles(backendDir, config)
	if backendGenErr != nil {
		return backendGenErr
	}

	mergeErr := fs.MergeDirs(workDir, append(getSourcePaths(reposDir, config), stateDir, backendDir))
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
	config, configErr := getConfig()
	if configErr != nil {
		fmt.Println(configErr.Error())
		os.Exit(1)
	}

	err := Exec(config)
	if err != nil {
		fmt.Println(err.Error())

		hookErr := config.TerminationHooks.Run(false)
		if hookErr != nil {
			fmt.Println(hookErr.Error())
		}

		os.Exit(1)
	}

	hookErr := config.TerminationHooks.Run(true)
	if hookErr != nil {
		fmt.Println(hookErr.Error())
		os.Exit(1)
	}
}
