package terraform

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path"

	"github.com/hashicorp/terraform-exec/tfexec"
)

func getPlanPath(dir string) string {
	return path.Join(dir, "terracd-plan")
}

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

func Plan(dir string, terraformPath string) (bool, error) {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return false, errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	changes, planErr := tf.Plan(context.Background(), tfexec.Out(getPlanPath(dir)))
	if planErr != nil {
		return false, errors.New(fmt.Sprintf("Error with terraform plan in directory \"%s\": %s", dir, planErr.Error()))
	}

	return changes, nil
}

func Apply(dir string, terraformPath string) error {
	tf, err := tfexec.NewTerraform(dir, terraformPath)
	if err != nil {
		return errors.New(fmt.Sprintf("Error preparing terraform in directory \"%s\": %s", dir, err.Error()))
	}

	tf.SetStdout(os.Stdout)
	tf.SetStderr(os.Stderr)

	applyErr := tf.Apply(context.Background(), tfexec.DirOrPlan(getPlanPath(dir)))
	if applyErr != nil {
		return errors.New(fmt.Sprintf("Error with terraform apply in directory \"%s\": %s", dir, applyErr.Error()))
	}

	return nil
}