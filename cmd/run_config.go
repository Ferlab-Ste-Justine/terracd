package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Ferlab-Ste-Justine/terracd/cache"
	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
	"github.com/Ferlab-Ste-Justine/terracd/state"
)

func backupState(workDir string, stateDir string) error {
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

	return nil
}

func cacheProviders(workDir string, cacheDir string) error {
	lockFileSrc := path.Join(workDir, ".terraform.lock.hcl")
	lockFileDest := path.Join(cacheDir, ".terraform.lock.hcl")
	providerDirSrc := path.Join(workDir, ".terraform", "providers")
	providerDirDest := path.Join(cacheDir, ".terraform", "providers")

	lockFileSrcExists, lockFileSrcErr := fs.PathExists(lockFileSrc)
	if lockFileSrcErr != nil {
		return lockFileSrcErr
	}

	providerDirSrcExists, providerDirSrcErr := fs.PathExists(providerDirSrc)
	if providerDirSrcErr != nil {
		return providerDirSrcErr
	}

	providerDirDestExists, providerDirDestErr := fs.PathExists(providerDirDest)
	if providerDirDestErr != nil {
		return providerDirDestErr
	}

	if lockFileSrcExists && providerDirSrcExists {
		if providerDirDestExists {
			remErr := os.RemoveAll(providerDirDest)
			if remErr != nil {
				return remErr
			}
		}
		
		copyErr := fs.CopyPrivateFile(lockFileSrc, lockFileDest)
		if copyErr != nil {
			return copyErr
		}

		ensErr := fs.AssurePrivateDir(providerDirDest)
		if ensErr != nil {
			return ensErr
		}

		copyErr = fs.CopyDir(providerDirDest, providerDirSrc)
		if copyErr != nil {
			return copyErr
		}
	}

	return nil
}

func cleanup(workDir string, stateDir string, cacheDir string, cache bool) error {
	backupErr := backupState(workDir, stateDir)
	if backupErr != nil {
		return backupErr
	}

	if cache {
		cacheErr := cacheProviders(workDir, cacheDir)
		if cacheErr != nil {
			return cacheErr
		}
	}

	return os.RemoveAll(workDir)
}

func RunConfig(paths fs.Paths, conf config.Config, st state.State) (state.State, bool, error) {
	workDirExists, workDirExistsErr := fs.PathExists(paths.Root)
	if workDirExistsErr != nil {
		return st, false, workDirExistsErr
	}
	if !workDirExists {
		assureErr := fs.AssurePrivateDir(paths.Root)
		if assureErr != nil {
			return st, false, assureErr
		}
	}

	chdirErr := os.Chdir(paths.Root)
	if chdirErr != nil {
		return st, false, chdirErr
	}

	workDirExists, workDirExistsErr = fs.PathExists(paths.Work)
	if workDirExistsErr != nil {
		return st, false, workDirExistsErr
	}
	if workDirExists {
		fmt.Println("Warning: Working directory found from prior iteration. Will clean it up.")
		cleanupErr := cleanup(paths.Work, paths.State, paths.Cache, false)
		if cleanupErr != nil {
			return st, false, cleanupErr
		}
	}

	assureErr := fs.AssurePrivateDir(paths.Repos)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Backend)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.State)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Cache)
	if assureErr != nil {
		return st, false, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Work)
	if assureErr != nil {
		return st, false, assureErr
	}

	defer func() {
		cleanupErr := cleanup(paths.Work, paths.State, paths.Cache, conf.Cache.IsDefined() && conf.Command != "wait")
		if cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup working directory at the end of execution: %s.\n", cleanupErr.Error())
		}
	}()

	commitHashes, syncErr := conf.Sources.SyncGitRepos(paths.Repos)
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

	backendGenErr := conf.Sources.GenerateBackendFiles(paths.Backend)
	if backendGenErr != nil {
		return st, false, backendGenErr
	}

	mergeDirs := append(conf.Sources.GetFsPaths(paths.Repos), paths.State, paths.Backend)
	mergeErr := fs.MergeDirs(paths.Work, mergeDirs)
	if mergeErr != nil {
		return st, false, mergeErr
	}

	var cacheInfo cache.CacheInfo
	var cacheInfoErr error
	if conf.Cache.IsDefined() {
		cacheInfo, cacheInfoErr = cache.GetCacheInfo(paths.Work, conf.Cache)
		if cacheInfoErr != nil {
			return st, false, cacheInfoErr
		}
	
		if cacheInfo.ShouldUse(&st.CacheInfo) {
			mergeErr := fs.MergeDirs(paths.Work, []string{paths.Cache})
			if mergeErr != nil {
				return st, false, mergeErr
			}
		}
	}

	switch conf.Command {
	case "wait":
		waitTime := conf.Timeouts.Wait
		if int64(waitTime) == int64(0) {
			waitTime, _ = time.ParseDuration("1h")
		}
		time.Sleep(waitTime)
	case "plan":
		planErr := Plan(paths.Work, conf)
		if planErr != nil {
			return st, false, planErr
		}
	case "apply":
		applyErr := Apply(paths.Work, conf)
		if applyErr != nil {
			return st, false, applyErr
		}
	case "migrate_backend":
		migrateErr := MigrateBackend(paths.Work, conf)
		if migrateErr != nil {
			return st, false, migrateErr
		}
	}

	return state.State{
		LastCommandOccurrence: *cmdOcc,
		CacheInfo: cacheInfo,
	}, false, nil
}