package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"

	"ferlab/terracd/git"
)

func getRepoDir(url string, ref string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s|%s", url, ref)))
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
		} else {
			paths = append(paths, source.Dir)
		}
	}

	return paths
}

func syncConfigRepos(dir string, c Config) error {
	for _, source := range c.Sources {
		if source.Repo.Url != "" {
			repoDir := path.Join(dir, getRepoDir(source.Repo.Url, source.Repo.Ref))
			
			_, err := os.Stat(repoDir)
			if err != nil {
				if !os.IsNotExist(err) {
					return errors.New(fmt.Sprintf("Error accessing repo directory \"%s\": %s", repoDir, err.Error()))
				}
		
				mkdirErr := os.MkdirAll(repoDir, 0770)
				if mkdirErr != nil {
					return errors.New(fmt.Sprintf("Error creating repo directory \"%s\": %s", repoDir, mkdirErr.Error()))
				}
			}
	
			err = git.SyncGitRepo(repoDir, source.Repo.Url, source.Repo.Ref, source.Repo.Auth.SshKeyPath, source.Repo.Auth.KnownHostsPath)
			if err != nil {
				return errors.New(fmt.Sprintf("Error updating branch \"%s\" of repo \"%s\": %s", source.Repo.Ref, source.Repo.Url, err.Error()))
			}
		}
	}

	return nil
}