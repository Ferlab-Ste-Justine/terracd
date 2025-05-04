package cmd

import (
	"fmt"
	"os"
	"path"
	"time"

	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/jitter"
	"github.com/Ferlab-Ste-Justine/terracd/metrics"
	"github.com/Ferlab-Ste-Justine/terracd/recurrence"
	"github.com/Ferlab-Ste-Justine/terracd/state"
)

func backupFsState(workDir string, stateDir string) error {
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

func cleanup(workDir string, stateDir string) error {
	backupErr := backupFsState(workDir, stateDir)
	if backupErr != nil {
		return backupErr
	}

	return os.RemoveAll(workDir)
}

func RunConfig(paths fs.Paths, conf config.Config, st state.State) (state.State, bool, []metrics.Provider, error) {
	fmt.Printf("Info: Running %s command.\n", conf.Command)
	
	workDirExists, workDirExistsErr := fs.PathExists(paths.Root)
	if workDirExistsErr != nil {
		return st, false, []metrics.Provider{}, workDirExistsErr
	}
	if !workDirExists {
		assureErr := fs.AssurePrivateDir(paths.Root)
		if assureErr != nil {
			return st, false, []metrics.Provider{}, assureErr
		}
	}

	workDirExists, workDirExistsErr = fs.PathExists(paths.Work)
	if workDirExistsErr != nil {
		return st, false, []metrics.Provider{}, workDirExistsErr
	}
	if workDirExists {
		fmt.Println("Warning: Working directory found from prior iteration. Will clean it up.")
		cleanupErr := cleanup(paths.Work, paths.State)
		if cleanupErr != nil {
			return st, false, []metrics.Provider{}, cleanupErr
		}
	}

	assureErr := fs.AssurePrivateDir(paths.Repos)
	if assureErr != nil {
		return st, false, []metrics.Provider{}, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Backend)
	if assureErr != nil {
		return st, false, []metrics.Provider{}, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.State)
	if assureErr != nil {
		return st, false, []metrics.Provider{}, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Cache)
	if assureErr != nil {
		return st, false, []metrics.Provider{}, assureErr
	}

	assureErr = fs.AssurePrivateDir(paths.Work)
	if assureErr != nil {
		return st, false, []metrics.Provider{}, assureErr
	}

	defer func() {		
		cleanupErr := cleanup(paths.Work, paths.State)
		if cleanupErr != nil {
			fmt.Printf("Warning: Failed to cleanup working directory at the end of execution: %s.\n", cleanupErr.Error())
		}
	}()

	commitHashes, syncErr := conf.Sources.SyncGitRepos(paths.Repos)
	if syncErr != nil {
		return st, false, []metrics.Provider{}, syncErr
	}

	cmdOcc := recurrence.GenerateCommandOccurrence(conf.Command, commitHashes)
	if conf.Recurrence.IsDefined() {
		if !st.LastCommandOccurrence.ShouldOccur(&conf.Recurrence, cmdOcc) {
			fmt.Println("Info: Recurrence policy dictates that execution should be skipped at this time.")
			return st, true, []metrics.Provider{}, nil
		}
	}

	backendGenErr := conf.Sources.GenerateBackendFiles(paths.Backend)
	if backendGenErr != nil {
		return st, false, []metrics.Provider{}, backendGenErr
	}

	mergeDirs := append(conf.Sources.GetFsPaths(paths.Repos), paths.State, paths.Backend)
	mergeErr := fs.MergeDirs(paths.Work, mergeDirs)
	if mergeErr != nil {
		return st, false, []metrics.Provider{}, mergeErr
	}

	cacheInfo, cacheDirInfo, cacheInfoErr := conf.Cache.Load(paths.Work, paths.Cache, st.CacheInfo)
	if cacheInfoErr != nil {
		return st, false, []metrics.Provider{}, cacheInfoErr
	}

	defer func() {
		if conf.Command != "wait" {
			saveCacheErr := conf.Cache.Save(paths.Work, paths.Cache, cacheDirInfo)
			if saveCacheErr != nil {
				fmt.Printf("Warning: Failed to backup cache: %s.\n", saveCacheErr.Error())
			}
		}
	}()

	if conf.RandomJitter > 0 {
		jitter.Seed()
		sleepDuration := jitter.GetRandomDuration(conf.RandomJitter)
		fmt.Printf("Info: Sleeping for %s\n", jitter.Stringify(sleepDuration))
		time.Sleep(sleepDuration)
	}

	previousWd, previousWdErr := os.Getwd()
	if previousWdErr != nil {
		return st, false, []metrics.Provider{}, previousWdErr
	}

	chdirErr := os.Chdir(paths.Root)
	if chdirErr != nil {
		return st, false, []metrics.Provider{}, chdirErr
	}

	defer func() {
		chdirErr := os.Chdir(previousWd)
		if chdirErr != nil {
			fmt.Printf("Warning: Failed reverting working directory after running terraform: %s\n", chdirErr.Error())
		}
	}()

	switch conf.Command {
	case "wait":
		waitTime := conf.Timeouts.Wait
		if int64(waitTime) == int64(0) {
			waitTime, _ = time.ParseDuration("1h")
		}
		time.Sleep(waitTime)
	case "plan":
		_, planErr := Plan(paths.Work, conf)
		if planErr != nil {
			return st, false, []metrics.Provider{}, planErr
		}
	case "apply":
		applied, applyErr := Apply(paths.Work, conf)
		if applyErr != nil {
			return st, false, []metrics.Provider{}, applyErr
		}
		if !applied {
			fmt.Println("Info: Plan indicated no operations. Skipped apply.")
		}
	case "destroy":
		destroyErr := Destroy(paths.Work, conf)
		if destroyErr != nil {
			return st, false, []metrics.Provider{}, destroyErr
		}
	case "migrate_backend":
		migrateErr := MigrateBackend(paths.Work, conf)
		if migrateErr != nil {
			return st, false, []metrics.Provider{}, migrateErr
		}
	}

	var usedProvidersErr error
	usedProviders := []metrics.Provider{}
	if conf.Command != "wait" && conf.Metrics.IncludeProviders {
		usedProviders, usedProvidersErr = metrics.GetProvidersInfo(paths.Work)
		if usedProvidersErr != nil {
			return state.State{
				LastCommandOccurrence: *cmdOcc,
				CacheInfo: cacheInfo,
			}, false, usedProviders, usedProvidersErr
		}
	}

	return state.State{
		LastCommandOccurrence: *cmdOcc,
		CacheInfo: cacheInfo,
	}, false, usedProviders, nil
}