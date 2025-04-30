package main

import (
	"os"
	"path"
	"testing"
	"time"

	etcdtest "github.com/Ferlab-Ste-Justine/etcd-sdk/testutils"
	git "github.com/Ferlab-Ste-Justine/git-sdk"
	gittest "github.com/Ferlab-Ste-Justine/git-sdk/testutils"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
)

func TestPlanSuccessFailureSkip(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "plan",
		MinInterval: "60s",
		Jitter: "2s",
		State: TestConfTemplateState{
			Type: "Fs",
		},
		DirSources: []TestConfTemplateDirSrc{
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}
	defer func() {
		err := CleanupTestExecution(tpl)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
		return
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected plan to succeed and it didn't")
		return
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
		return
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of plan to skip and it didn't")
		return
	}

	tpl.MinInterval = "1ms"
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
		return
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected third iteration of plan to succeed and it didn't")
		return
	}

	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
		return
	}

	if hooks4.Success != hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure == time.Duration(0) {
		t.Errorf("Expected fourth iteration of plan to fail and it didn't")
		return
	}
}

func TestApplySuccessFailureSkip(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "apply",
		MinInterval: "60s",
		Jitter: "2s",
		State: TestConfTemplateState{
			Type: "Fs",
		},
		DirSources: []TestConfTemplateDirSrc{
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}
	defer func() {
		err := CleanupTestExecution(tpl)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
		return
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr := FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "A")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected file to have a value of 'A' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
		return
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of apply to skip and it didn't")
		return
	}

	tpl.MinInterval = "1ms"
	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValB")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
		return
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected third iteration of apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "B")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After third iteration, expected file to have a value of 'B' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After third iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
		return
	}

	if hooks4.Success != hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure == time.Duration(0) {
		t.Errorf("Expected fourth iteration of apply to fail and it didn't")
		return
	}
}

func TestDestroySuccessFailureSkip(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "apply",
		MinInterval: "60s",
		Jitter: "2s",
		State: TestConfTemplateState{
			Type: "Fs",
		},
		DirSources: []TestConfTemplateDirSrc{
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}
	defer func() {
		err := CleanupTestExecution(tpl)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
		return
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr := FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "A")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected file to have a value of 'A' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	tpl.Command = "destroy"

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()

	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
		return
	}

	if hooks2.Success == hooks.Success || hooks2.Skip != time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected destroy to succeed and it didn't")
		return
	}

	exists, existsErr := fs.PathExists(path.Join("e2e_test", "runtime", "output", "file"))
	if existsErr != nil {
		t.Errorf(existsErr.Error())
		return
	}

	if exists {
		t.Errorf("Expected file to be destroyed and it wasn't")
		return
	}

	exists, existsErr = fs.PathExists(path.Join("e2e_test", "runtime", "output", "file-other"))
	if existsErr != nil {
		t.Errorf(existsErr.Error())
		return
	}

	if exists {
		t.Errorf("Expected other file to be destroyed and it wasn't")
		return
	}

	tpl.MinInterval = "1ms"
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()

	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
		return
	}

	if hooks3.Success != hooks2.Success || hooks3.Skip == time.Duration(0) || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected destroy to skip and it didn't")
		return
	}

	tpl.Command = "plan"
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()

	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
		return
	}

	if hooks4.Success == hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure != time.Duration(0) {
		t.Errorf("Expected plan to succeed and it didn't")
		return
	}

	tpl.Command = "destroy"
	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()

	hooks5, hooks5Err := GetTestHooks()
	if hooks5Err != nil {
		t.Errorf(hooks5Err.Error())
		return
	}

	if hooks5.Success != hooks4.Success || hooks5.Skip != hooks4.Skip || hooks5.Failure == time.Duration(0) {
		t.Errorf("Expected destroy to fail and it didn't")
		return
	}
}

func TestEtcdState(t *testing.T) {
	tearDown, launchErr := etcdtest.LaunchTestEtcdCluster(path.Join("e2e_test", "etcd-dependencies"), etcdtest.EtcdTestClusterOpts{})
	if launchErr != nil {
		t.Errorf("Error occured launching test etcd cluster: %s", launchErr.Error())
		return
	}

	defer func() {
		errs := tearDown()
		if len(errs) > 0 {
			t.Errorf("Errors occured tearing down etcd cluster: %s", errs[0].Error())
		}
	}()

	tpl := TestConfTemplate{
		Command: "apply",
		MinInterval: "60s",
		Jitter: "2s",
		State: TestConfTemplateState{
			Type: "Etcd",
		},
		DirSources: []TestConfTemplateDirSrc{
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}
	defer func() {
		err := CleanupTestExecution(tpl)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
		return
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr := FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "A")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected file to have a value of 'A' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
		return
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of apply to skip and it didn't")
		return
	}

	tpl.MinInterval = "1ms"
	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValB")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
		return
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected third iteration of apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "B")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After third iteration, expected file to have a value of 'B' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After third iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	tpl.DirSources = []TestConfTemplateDirSrc{
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	MainNoExit()
	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
		return
	}

	if hooks4.Success != hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure == time.Duration(0) {
		t.Errorf("Expected fourth iteration of apply to fail and it didn't")
		return
	}
}

func TestGitApplyRecurrence(t *testing.T) {
    sshPub, sshPubErr := os.ReadFile(path.Join("e2e_test", "git-dependencies", "keys", "ssh", "id_rsa.pub"))
	if sshPubErr != nil {
		t.Errorf("Error occured reading user's public ssh key: %s", sshPubErr.Error())
		return
	}
	
	workDir, workDirErr := os.Getwd()
	if workDirErr != nil {
		t.Errorf("Error getting current working directory: %s", workDirErr.Error())
		return
	}

	giteaPath := path.Join(workDir, "gitea")
	giteaPathErr := os.MkdirAll(giteaPath, 0770)
	if giteaPathErr != nil {
		t.Errorf("Error creating gitea working directory: %s", giteaPathErr.Error())
		return
	}
	defer func() {
		err := os.RemoveAll(giteaPath)
		if err != nil {
			t.Errorf("Error cleaning up gitea working directory: %s", err.Error())
		}
	} ()

	giteaTeardown, giteaInfo, giteaErr := gittest.LaunchTestGitea(gittest.GiteaOpts{
		Workdir: giteaPath,
		BindIp: "127.0.0.1",
		BindPort: 3000,
		Password: "test",
		Email: "test@test.test",
		SshPub: string(sshPub),
		Repos: []string{"local_file"},
		Debug: true,
	})
	if giteaErr != nil {
		t.Errorf("Error starting gitea: %s", giteaErr.Error())
		return
	}
	defer func() {
		err := giteaTeardown()
		if err != nil {
			t.Errorf("Error stopping gitea: %s", err.Error())
		}
	} ()

	sshCreds, sshCredsErr := git.GetSshCredentials(path.Join("e2e_test", "git-dependencies", "keys", "ssh", "id_rsa"), giteaInfo.KnownHostsFile, giteaInfo.User)
	if sshCredsErr != nil {
		t.Errorf("Error retrieving ssh credentials: %s", sshCredsErr.Error())
		return
	}

	workRepoPath := path.Join(workDir, "work_repo")
	workRepoErr := os.MkdirAll(workRepoPath, 0770)
	if workRepoErr != nil {
		t.Errorf("Error creating repo working directory: %s", workRepoErr.Error())
		return
	}
	defer func() {
		err := os.RemoveAll(workRepoPath)
		if err != nil {
			t.Errorf("Error cleaning up repo working directory: %s", err.Error())
		}
	} ()

	setErr := setRepoLocalFileContent("test", workRepoPath, giteaInfo.RepoUrls[0], sshCreds, giteaInfo.User)
	if setErr != nil {
		t.Errorf("Error setting file value in gitea repo: %s", setErr.Error())
		return
	}

	tpl := TestConfTemplate{
		Command: "apply",
		MinInterval: "120s",
		Jitter: "2s",
		State: TestConfTemplateState{
			Type: "Fs",
		},
		DirSources: []TestConfTemplateDirSrc{
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateDirSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
		GitSources: []TestConfTemplateGitSrc{
			TestConfTemplateGitSrc{
				Url: giteaInfo.RepoUrls[0],
				User: giteaInfo.User,
				KnownHost: giteaInfo.KnownHostsFile,
			},
		},
	}
	defer func() {
		err := CleanupTestExecution(tpl)
		if err != nil {
			t.Errorf(err.Error())
		}
	}()

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
		return
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
		return
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
		return
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr := FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "A")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected file to have a value of 'A' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected other file to have a value of 'O' after apply and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "git_file"), "test")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After first iteration, expected git file to have a value of 'test' after apply and it didn't")
		return
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
		return
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of apply to skip and it didn't")
		return
	}

	setErr = setRepoLocalFileContent("test2", workRepoPath, giteaInfo.RepoUrls[0], sshCreds, giteaInfo.User)
	if setErr != nil {
		t.Errorf("Error setting file value in gitea repo: %s", setErr.Error())
		return
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
		return
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
		return
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "git_file"), "test2")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
		return
	}

	if !hasVal {
		t.Errorf("After third iteration, expected git file to have a value of 'test2' after apply and it didn't")
		return
	}	
}