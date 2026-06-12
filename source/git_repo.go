package source

import (
	"encoding/base64"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"

	yaml "gopkg.in/yaml.v2"
	git "github.com/Ferlab-Ste-Justine/git-sdk"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
)

type CommitHash struct {
	Url  string
	Ref  string
	Path string
	Hash string 
}

type GitRepoAuthSsh struct {
	SshKeyPath     string `yaml:"ssh_key_path"`
	KnownHostsPath string `yaml:"known_hosts_path"`
	User  string
}

func (auth *GitRepoAuthSsh) IsDefined() bool {
	return auth.SshKeyPath != ""
}

type BasicAuth struct {
	Username string
	Password string
}

type GitRepoAuthHttps struct {
	BasicAuthPath string `yaml:"basic_auth_path"`
}

func (auth *GitRepoAuthHttps) IsDefined() bool {
	return auth.BasicAuthPath != ""
}

func (auth *GitRepoAuthHttps) GetBasicAuth() (BasicAuth, error) {
	var bAuth BasicAuth

	b, err := ioutil.ReadFile(auth.BasicAuthPath)
	if err != nil {
		return bAuth, errors.New(fmt.Sprintf("Error reading the basic auth file: %s", err.Error()))
	}
	
	err = yaml.Unmarshal(b, &bAuth)
	if err != nil {
		return bAuth, errors.New(fmt.Sprintf("Error parsing the basic auth file: %s", err.Error()))
	}

	return bAuth, nil
}

type GitRepoAuth struct {
	Ssh   GitRepoAuthSsh
	Https GitRepoAuthHttps
}

func (auth *GitRepoAuth) IsDefined() bool {
	return auth.Ssh.IsDefined() || auth.Https.IsDefined()
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

func (repo *GitRepo) Sync(dir string) (CommitHash, error) {
	repoDir := path.Join(dir, repo.GetDir())

	_, err := os.Stat(repoDir)
	if err != nil {
		if !os.IsNotExist(err) {
			return CommitHash{}, errors.New(fmt.Sprintf("Error accessing repo directory \"%s\": %s", repoDir, err.Error()))
		}

		mkdirErr := os.MkdirAll(repoDir, 0770)
		if mkdirErr != nil {
			return CommitHash{}, errors.New(fmt.Sprintf("Error creating repo directory \"%s\": %s", repoDir, mkdirErr.Error()))
		}
	}

	armoredKeyrings := []string{}
	for _, armoredKeyRingPath := range repo.GpgPublicKeysPaths {
		armoredKeyRingFiles, err := fs.FindFiles(armoredKeyRingPath, "*")
		if err != nil {
			return CommitHash{}, errors.New(fmt.Sprintf("Error finding armored keyring files at \"%s\": %s", armoredKeyRingPath, err.Error()))
		}

		for _, armoredKeyRingFile := range armoredKeyRingFiles {
			armoredKeyring, err := os.ReadFile(armoredKeyRingFile)
			if err != nil {
				return CommitHash{}, errors.New(fmt.Sprintf("Error reading armored keyring \"%s\": %s", armoredKeyRingFile, err.Error()))
			}
			armoredKeyrings = append(armoredKeyrings, string(armoredKeyring))
		}
	}

	var gitCreds *git.GitCredentials
	var gitCredsErr error
	if repo.Auth.IsDefined() {
		if repo.Auth.Ssh.IsDefined() {
			gitCreds, gitCredsErr = git.GetSshCredentials(repo.Auth.Ssh.SshKeyPath, repo.Auth.Ssh.KnownHostsPath, repo.Auth.Ssh.User)
			if gitCredsErr != nil {
				return CommitHash{}, gitCredsErr
			}
		} else {
			basicAuth, basicAuthErr := repo.Auth.Https.GetBasicAuth()
			if basicAuthErr != nil {
				return CommitHash{}, basicAuthErr
			}

			gitCreds = git.GetHttpsCredentials(basicAuth.Username, basicAuth.Password)
		}
	}

	gogitRepo, badRepoDir, syncErr := git.SyncGitRepo(repoDir, repo.Url, repo.Ref, gitCreds)
	if syncErr != nil {
		if !badRepoDir {
			return CommitHash{}, errors.New(fmt.Sprintf("Error updating branch \"%s\" of repo \"%s\": %s", repo.Ref, repo.Url, syncErr.Error()))
		}

		fmt.Printf("Warning: Will delete repo dir to circumvent error: %s\n", syncErr.Error())

		removalErr := os.RemoveAll(repoDir)
		if removalErr != nil {
			return CommitHash{}, removalErr
		}

		return repo.Sync(dir)
	}

	if len(armoredKeyrings) > 0 {
		verErr := git.VerifyTopCommit(gogitRepo, armoredKeyrings)
		if verErr != nil {
			return CommitHash{}, verErr
		}
	}

	head, headErr := gogitRepo.Repo.Head()
	if headErr != nil {
		return CommitHash{}, headErr
	}

	return CommitHash{
		Url: repo.Url,
		Ref: repo.Ref,
		Path: repo.Path,
		Hash: head.Hash().String(),
	}, nil
}

func (srcs *Sources) SyncGitRepos(dir string) ([]CommitHash, error) {
	hashes := []CommitHash{}
	for _, source := range *srcs {
		if source.GetType() == TypeGitRepo {
			hash, err := source.GitRepo.Sync(dir)
			if err != nil {
				return hashes, err
			}
			hashes = append(hashes, hash)
		}
	}

	return hashes, nil
}