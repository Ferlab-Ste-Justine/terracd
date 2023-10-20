package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
	"github.com/Ferlab-Ste-Justine/terracd/state"
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

func RunConfig(conf config.Config, st state.State) (state.State, bool, error) {
	workDirExists, workDirExistsErr := fs.PathExists(conf.WorkingDirectory)
	if workDirExistsErr != nil {
		return st, false, workDirExistsErr
	}
	if !workDirExists {
		assureErr := fs.AssurePrivateDir(conf.WorkingDirectory)
		if assureErr != nil {
			return st, false, assureErr
		}
	}

	chdirErr := os.Chdir(conf.WorkingDirectory)
	if chdirErr != nil {
		return st, false, chdirErr
	}

	reposDir := path.Join(conf.WorkingDirectory, "repos")
	backendDir := path.Join(conf.WorkingDirectory, "backend")
	stateDir := path.Join(conf.WorkingDirectory, "state")
	workDir := path.Join(conf.WorkingDirectory, "work")

	workDirExists, workDirExistsErr = fs.PathExists(workDir)
	if workDirExistsErr != nil {
		return st, false, workDirExistsErr
	}
	if workDirExists {
		fmt.Println("Warning: Working directory found from prior iteration. Will clean it up.")
		cleanupErr := cleanup(workDir, stateDir)
		if cleanupErr != nil {
			return st, false, cleanupErr
		}
	}

	assureErr := fs.AssurePrivateDir(reposDir)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(backendDir)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(stateDir)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(workDir)
	if assureErr != nil {
		return st, false, assureErr
	}

	defer func() {
		cleanupErr := cleanup(workDir, stateDir)
		if cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup working directory at the end of execution: %s.\n", cleanupErr.Error())
		}
	}()

	commitHashes, syncErr := conf.Sources.SyncGitRepos(reposDir)
	if syncErr != nil {
		return st, false, syncErr
	}

	cmdOcc := recurrence.GenerateCommandOccurrence(conf.Command, commitHashes)
	if conf.Recurrence.IsDefined() {
		if !st.LastCommandOccurrence.ShouldOccur(&conf.Recurrence, cmdOcc) {
			fmt.Printf("Info: Recurrence policy dictates that execution should be skipped at this time.")
			return st, true, nil
		}
	}

	backendGenErr := conf.Sources.GenerateBackendFiles(backendDir)
	if backendGenErr != nil {
		return st, false, backendGenErr
	}

	mergeErr := fs.MergeDirs(workDir, append(conf.Sources.GetFsPaths(reposDir), stateDir, backendDir))
	if mergeErr != nil {
		return st, false, mergeErr
	}

	switch conf.Command {
	case "wait":
		waitTime := conf.Timeouts.Wait
		if int64(waitTime) == int64(0) {
			waitTime, _ = time.ParseDuration("1h")
		}
		time.Sleep(waitTime)
	case "plan":
		planErr := Plan(workDir, conf)
		if planErr != nil {
			return st, false, planErr
		}
	case "apply":
		applyErr := Apply(workDir, conf)
		if applyErr != nil {
			return st, false, applyErr
		}
	case "migrate_backend":
		migrateErr := MigrateBackend(workDir, conf)
		if migrateErr != nil {
			return st, false, migrateErr
		}
	}

	return state.State{
		LastCommandOccurrence: *cmdOcc,
	}, false, nil
}