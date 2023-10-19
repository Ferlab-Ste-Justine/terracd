package source

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"
	"path"

	git "github.com/Ferlab-Ste-Justine/git-sdk"

	"ferlab/terracd/fs"
)

type GitRepoAuth struct {
	SshKeyPath     string `yaml:"ssh_key_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
}

type GitRepo struct {
	Url                string
	Ref                string
	Path               string
	Auth               GitRepoAuth
	GpgPublicKeysPaths []string `yaml:"gpg_public_keys_paths"`
}

func (repo *GitRepo) GetDir() string {
	return base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("%s|%s", repo.Url, repo.Ref)))
}

func (repo *GitRepo) Sync(dir string) error {
	repoDir := path.Join(dir, repo.GetDir())

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
	for _, armoredKeyRingPath := range repo.GpgPublicKeysPaths {
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

	sshCreds, sshCredsErr := git.GetSshCredentials(repo.Auth.SshKeyPath, repo.Auth.KnownHostsPath)
	if sshCredsErr != nil {
		return sshCredsErr
	}

	gogitRepo, badRepoDir, syncErr := git.SyncGitRepo(repoDir, repo.Url, repo.Ref, sshCreds)
	if syncErr != nil {
		if !badRepoDir {
			return errors.New(fmt.Sprintf("Error updating branch \"%s\" of repo \"%s\": %s", repo.Ref, repo.Url, syncErr.Error()))
		}

		fmt.Printf("Warning: Will delete repo dir to circumvent error: %s\n", syncErr.Error())

		removalErr := os.RemoveAll(repoDir)
		if removalErr != nil {
			return removalErr
		}

		return repo.Sync(dir)
	}

	if len(armoredKeyrings) > 0 {
		return git.VerifyTopCommit(gogitRepo, armoredKeyrings)
	}

	return nil
}

func (srcs *Sources) SyncGitRepos(dir string) error {
	for _, source := range *srcs {
		if source.GetType() == TypeGitRepo {
			err := source.GitRepo.Sync(dir)
			if err != nil {
				return err
			}
		}
	}

	return nil
}