package main

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"

	"ferlab/terracd/fs"

	git "github.com/Ferlab-Ste-Justine/git-sdk"
)

func getRepoDir(url string, ref string) string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s|%s", url, ref)))
}

func syncConfigRepo(dir string, source ConfigSource, c Config) error {
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

	armoredKeyrings := []string{}
	for _, armoredKeyRingPath := range source.Repo.GpgPublicKeysPaths {
		armoredKeyRingFiles, err := fs.FindFiles(armoredKeyRingPath, "*")
		if err != nil {
			return errors.New(fmt.Sprintf("Error finding armored keyring files at \"%s\": %s", armoredKeyRingPath, err.Error()))
		}

		for _, armoredKeyRingFile := range armoredKeyRingFiles {
			armoredKeyring, err := os.ReadFile(armoredKeyRingFile)
			if err != nil {
				return errors.New(fmt.Sprintf("Error reading armored keyring \"%s\": %s", armoredKeyRingFile, err.Error()))
			}
			armoredKeyrings = append(armoredKeyrings, string(armoredKeyring))
		}
	}

	sshCreds, sshCredsErr := git.GetSshCredentials(source.Repo.Auth.SshKeyPath, source.Repo.Auth.KnownHostsPath)
	if sshCredsErr != nil {
		return sshCredsErr
	}

	repo, badRepoDir, syncErr := git.SyncGitRepo(repoDir, source.Repo.Url, source.Repo.Ref, sshCreds)
	if syncErr != nil {
		if !badRepoDir {
			return errors.New(fmt.Sprintf("Error updating branch \"%s\" of repo \"%s\": %s", source.Repo.Ref, source.Repo.Url, syncErr.Error()))
		}

		fmt.Printf("Warning: Will delete repo dir to circumvent error: %s\n", syncErr.Error())

		removalErr := os.RemoveAll(repoDir)
		if removalErr != nil {
			return removalErr
		}

		return syncConfigRepo(dir, source, c)
	}

	if len(armoredKeyrings) > 0 {
		return git.VerifyTopCommit(repo, armoredKeyrings)
	}

	return nil
}

func syncConfigRepos(dir string, c Config) error {
	for _, source := range c.Sources {
		if source.Repo.Url != "" {
			err := syncConfigRepo(dir, source, c)
			if err != nil {
				return err
			}
		}
	}

	return nil
}
