package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"
	"time"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func getContext(timeout time.Duration) (context.Context, context.CancelFunc) {
	ctx := context.Background()
	if int64(timeout) == int64(0) {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, timeout)
}

func Init(dir string, terraformPath string, timeout time.Duration) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	ctx, cancel := getContext(timeout)
	defer cancel()

	initErr := tf.Init(ctx, tfexec.Upgrade(true))
	if initErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform init in directory \"%s\": %s", dir, initErr.Error()))
	}

	return nil
}

func Plan(dir string, planName string, terraformPath string, timeout time.Duration) (bool, error) {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	ctx, cancel := getContext(timeout)
	defer cancel()

	changes, planErr := tf.Plan(ctx, tfexec.Out(path.Join(dir, planName)))
	if planErr != nil {
		return false, errors.New(fmt.Sprintf("Error with terraform plan in directory \"%s\": %s", dir, planErr.Error()))
	}

	return changes, nil
}

func Apply(dir string, planName string, terraformPath string, timeout time.Duration) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	ctx, cancel := getContext(timeout)
	defer cancel()

	applyErr := tf.Apply(ctx, tfexec.DirOrPlan(path.Join(dir, planName)))
	if applyErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform apply in directory \"%s\": %s", dir, applyErr.Error()))
	}

	return nil
}

func operationsInsersect(a tfjson.Actions, b tfjson.Actions) bool {
	for _, aElem := range a {
		for _, bElem := range b {
			if aElem == bElem {
				return true
			}
		}
	}

	return false
}

func CheckPlan(dir string, planName string, terraformPath string, forbiddenOps []ForbiddenOperation) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	plan, planErr := tf.ShowPlanFile(context.Background(), path.Join(dir, planName))
	if planErr != nil {
		return errors.New(fmt.Sprintf("Error occured while reading/parsing the plan file: %s", planErr.Error()))
	}

	for _, change := range plan.ResourceChanges {
		for _, forOp := range forbiddenOps {
			sameProvider := forOp.Provider == "" || forOp.Provider == change.ProviderName
			sameAddress := forOp.ResourceAddress == change.Address
			if sameProvider && sameAddress && operationsInsersect(forOp.Operations, (*change.Change).Actions) {
				return errors.New(fmt.Sprintf("Aborting as forbidden operation is about to be performed on protected resource \"%s\"", forOp.ResourceAddress))
			}
		}
	}

	return nil
}

func StatePull(dir string, stateFile string, terraformPath string, timeout time.Duration) (error) {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	ctx, cancel := getContext(timeout)
	defer cancel()

	state, pullErr := tf.StatePull(ctx)
	if pullErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform state pull in directory \"%s\": %s", dir, pullErr.Error()))
	}

	f, openErr := os.OpenFile(stateFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0700)
	if openErr != nil {
		return errors.New(fmt.Sprintf("Error opening file \"%s\" to store the terraform state: %s", stateFile, openErr.Error()))
	}

	defer f.Close()

	_, wErr := f.WriteString(state)
	if wErr != nil {
		return errors.New(fmt.Sprintf("Error writing to file \"%s\" to store the terraform state: %s", stateFile, wErr.Error()))
	}

	return nil
}

func StatePush(dir string, stateFile string, terraformPath string, timeout time.Duration) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	ctx, cancel := getContext(timeout)
	defer cancel()

	pushErr := tf.StatePush(ctx, stateFile)
	if pushErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform state push with state file \"%s\": %s", stateFile, pushErr.Error()))
	}

	return nil
}