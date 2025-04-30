package main

import (
	"bytes"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"text/template"
	"time"

	git "github.com/Ferlab-Ste-Justine/git-sdk"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
)

type TestConfTemplateDirSrc struct {
	Dir string
}

type TestConfTemplateGitSrc struct {
	Url string
	User string
	KnownHost string
}

type TestConfTemplateState struct {
	Type string
}

type TestConfTemplate struct {
	TerraformPath string
	Command       string
	MinInterval   string
	Jitter        string
	State         TestConfTemplateState
	DirSources    []TestConfTemplateDirSrc
	GitSources    []TestConfTemplateGitSrc
}

func (tpl *TestConfTemplate) SetTfPath() error {
	tfPath, tfPathErr := exec.LookPath("terraform")
	if tfPathErr != nil {
		return tfPathErr
	}

	tpl.TerraformPath = tfPath

	return nil
}

func (tpl *TestConfTemplate) GenerateConfig() error {
	fContent, err := ioutil.ReadFile(path.Join("e2e_test", "config.yml.tpl"))
	if err != nil {
		return err
	}

	tmpl, tErr := template.New("template").Parse(string(fContent))
	if tErr != nil {
		return tErr
	}

	var b bytes.Buffer
	exErr := tmpl.Execute(&b, tpl)
	if exErr != nil {
		return exErr
	}

	return os.WriteFile("config.yml", b.Bytes(), 0640)
}

func (tpl *TestConfTemplate) CleanupConfig() error {
	return os.Remove("config.yml")
}

type TestHooks struct {
	Success time.Duration
	Skip    time.Duration
	Failure time.Duration
}

func getTestHook(filePath string) (time.Duration, error) {
	pathExists, pathExistsErr := fs.PathExists(filePath)
	if pathExistsErr != nil {
		return time.Duration(0), pathExistsErr
	}

	if !pathExists {
		return time.Duration(0), nil
	}

	fContent, rErr := ioutil.ReadFile(filePath)
	if rErr != nil {
		return time.Duration(0), rErr
	}

	i, iErr := strconv.ParseInt(strings.TrimSpace(string(fContent)), 10, 64)
	if iErr != nil {
		return time.Duration(0), iErr
	}

	return time.Duration(i*1000000), nil
}

func GetTestHooks() (TestHooks, error) {
	f, fErr := getTestHook("failure")
	if fErr != nil {
		return TestHooks{}, fErr
	}

	sk, skErr := getTestHook("skip")
	if skErr != nil {
		return TestHooks{}, skErr
	}

	su, suErr := getTestHook("success")
	if suErr != nil {
		return TestHooks{}, suErr
	}

	return TestHooks{su, sk, f}, nil
}

func CleanupTestHooks() error {
	err := fs.EnsureFileNotExists("failure")
	if err != nil {
		return err
	}

	err = fs.EnsureFileNotExists("skip")
	if err != nil {
		return err
	}

	return fs.EnsureFileNotExists("success")
}

func CleanupTestRuntime() error {
	fExists, fExistsErr := fs.PathExists(path.Join("e2e_test", "runtime"))
	if fExistsErr != nil {
		return fExistsErr
	}

	if fExists {
		return os.RemoveAll(path.Join("e2e_test", "runtime"))
	}

	return nil
}

func CleanupTestExecution(tpl TestConfTemplate) error {
	err := tpl.CleanupConfig()
	if err != nil {
		return err
	}

	err = CleanupTestHooks()
	if err != nil {
		return err
	}

	return CleanupTestRuntime()
}

func FileHasValue(filePath string, value string) (bool, error) {
	pathExists, pathExistsErr := fs.PathExists(filePath)
	if pathExistsErr != nil {
		return false, pathExistsErr
	}

	if !pathExists {
		return false, nil
	}

	fContent, rErr := ioutil.ReadFile(filePath)
	if rErr != nil {
		return false, rErr
	}

	return string(fContent) == value, nil
}

func setRepoLocalFileContent(content string, fsPath string, repoUrl string, sshCreds *git.SshCredentials, user string) error {
	oneMinute, _ := time.ParseDuration("1m")
	return git.PushChanges(func() (*git.GitRepository, error) {
		repo, _, syncErr := git.SyncGitRepo(fsPath, repoUrl, "main", sshCreds)
		if syncErr != nil {
			return nil, syncErr
		}

		hclBit := `
		resource "local_file" "git_file" {
		content         = file("${path.module}/git_file_content.txt")
		file_permission = "0660"
		filename        = "${path.module}/../output/git_file"
		}
		`
	
		writeErr := os.WriteFile(path.Join(fsPath, "git_file.tf"), []byte(hclBit), 0644)
		if writeErr != nil {
			return nil, writeErr
		}
	
		writeErr = os.WriteFile(path.Join(fsPath, "git_file_content.txt"), []byte(content), 0644)
		if writeErr != nil {
			return nil, writeErr
		}

		_, commitErr := git.CommitFiles(repo, []string{"git_file.tf", "git_file_content.txt"}, "Some changes", git.CommitOptions{
			Name: user,
			Email: "test@test.test",
		})
		if commitErr != nil {
			return repo, commitErr
		}

		return repo, nil
	}, "main", sshCreds, 3, oneMinute)
}