package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/terraform-exec/tfexec"
	tfjson "github.com/hashicorp/terraform-json"
)

func Init(dir string, terraformPath string) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	initErr := tf.Init(context.Background(), tfexec.Upgrade(true))
	if initErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform init in directory \"%s\": %s", dir, initErr.Error()))
	}

	return nil
}

func Plan(dir string, planName string, terraformPath string) (bool, error) {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	changes, planErr := tf.Plan(context.Background(), tfexec.Out(path.Join(dir, planName)))
	if planErr != nil {
		return false, errors.New(fmt.Sprintf("Error with terraform plan in directory \"%s\": %s", dir, planErr.Error()))
	}

	return changes, nil
}

func Apply(dir string, planName string, terraformPath string) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	applyErr := tf.Apply(context.Background(), tfexec.DirOrPlan(path.Join(dir, planName)))
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