package main

import (
	"path"
	"testing"
	"time"
)

func TestPlanSuccessFailureSkip(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "plan",
		MinInterval: "60s",
		Jitter: "2s",
		Sources: []TestConfTemplateSrc{
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected plan to succeed and it didn't")
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of plan to skip and it didn't")
	}

	tpl.MinInterval = "1ms"
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected third iteration of plan to succeed and it didn't")
	}

	tpl.Sources = []TestConfTemplateSrc{
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}

	MainNoExit()
	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
	}

	if hooks4.Success != hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure == time.Duration(0) {
		t.Errorf("Expected fourth iteration of plan to fail and it didn't")
	}

	err = CleanupTestExecution(tpl)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestApply(t *testing.T) {
	tpl := TestConfTemplate{
		Command: "apply",
		MinInterval: "60s",
		Jitter: "2s",
		Sources: []TestConfTemplateSrc{
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileValA")},
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
			TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "version")},
		},
	}

	err := tpl.SetTfPath()
	if err != nil {
		t.Errorf(err.Error())
	}

	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}
	
	MainNoExit()
	
	hooks, hooksErr := GetTestHooks()
	if hooksErr != nil {
		t.Errorf(hooksErr.Error())
	}

	if hooks.Success == time.Duration(0) || hooks.Skip != time.Duration(0) || hooks.Failure != time.Duration(0) {
		t.Errorf("Expected apply to succeed and it didn't")
	}

	hasVal, hasValErr := FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "A")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
	}

	if !hasVal {
		t.Errorf("After first iteration, expected file to have a value of 'A' after apply and it didn't")
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
	}

	if !hasVal {
		t.Errorf("After first iteration, expected other file to have a value of 'O' after apply and it didn't")
	}

	MainNoExit()
	hooks2, hooks2Err := GetTestHooks()
	if hooks2Err != nil {
		t.Errorf(hooks2Err.Error())
	}

	if hooks.Success != hooks2.Success || hooks2.Skip == time.Duration(0) || hooks2.Failure != time.Duration(0) {
		t.Errorf("Expected second iteration of apply to skip and it didn't")
	}

	tpl.MinInterval = "1ms"
	tpl.Sources = []TestConfTemplateSrc{
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileValB")},
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileOther")},
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}

	MainNoExit()
	hooks3, hooks3Err := GetTestHooks()
	if hooks3Err != nil {
		t.Errorf(hooks3Err.Error())
	}

	if hooks3.Success == hooks2.Success || hooks3.Skip != hooks2.Skip || hooks3.Failure != time.Duration(0) {
		t.Errorf("Expected third iteration of apply to succeed and it didn't")
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file"), "B")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
	}

	if !hasVal {
		t.Errorf("After third iteration, expected file to have a value of 'B' after apply and it didn't")
	}

	hasVal, hasValErr = FileHasValue(path.Join("e2e_test", "runtime", "output", "file-other"), "O")
	if hasValErr != nil {
		t.Errorf(hasValErr.Error())
	}

	if !hasVal {
		t.Errorf("After third iteration, expected other file to have a value of 'O' after apply and it didn't")
	}

	tpl.Sources = []TestConfTemplateSrc{
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "fileBadSyntax")},
		TestConfTemplateSrc{Dir: path.Join("e2e_test", "tf", "version")},
	}
	err = tpl.GenerateConfig()
	if err != nil {
		t.Errorf(err.Error())
	}

	MainNoExit()
	hooks4, hooks4Err := GetTestHooks()
	if hooks4Err != nil {
		t.Errorf(hooks4Err.Error())
	}

	if hooks4.Success != hooks3.Success || hooks4.Skip != hooks3.Skip || hooks4.Failure == time.Duration(0) {
		t.Errorf("Expected fourth iteration of apply to fail and it didn't")
	}

	err = CleanupTestExecution(tpl)
	if err != nil {
		t.Errorf(err.Error())
	}
}

func TestDestroy(t *testing.T) {

}