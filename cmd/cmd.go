package cmd

import (
	"os"
	"path"

	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/terraform"
)

func MigrateBackend(dir string, conf config.Config) error {
	initErr := terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return initErr
	}

	currentBackend := path.Join(dir, conf.BackendMigration.CurrentBackend)
	rmErr := os.Remove(currentBackend)
	if rmErr != nil {
		return rmErr
	}

	copyErr := fs.CopyPrivateFile(conf.BackendMigration.NextBackend, path.Join(dir, path.Base(conf.BackendMigration.NextBackend)))
	if copyErr != nil {
		return copyErr
	}

	initErr = terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return initErr
	}

	return nil
}

func Plan(dir string, conf config.Config) (bool, error) {
	planName := "terracd-plan"
	forbiddenOpsFsPattern := "*.terracd-fo.yml"

	initErr := terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return false, initErr
	}

	changes, planErr := terraform.Plan(dir, planName, conf.TerraformPath, conf.Timeouts.TerraformPlan)
	if planErr != nil {
		return false, planErr
	}

	if !changes {
		return false, nil
	}

	forbiddenOpsFiles, foFilesErr := fs.FindFiles(dir, forbiddenOpsFsPattern)
	if foFilesErr != nil {
		return true, foFilesErr
	}

	forbiddenOps, foErr := terraform.GetForbiddenOperations(forbiddenOpsFiles)
	if foErr != nil {
		return true, foErr
	}

	return true, terraform.CheckPlan(dir, planName, conf.TerraformPath, forbiddenOps)
}

func Apply(dir string, conf config.Config) (bool, error) {
	planName := "terracd-plan"

	changes, planErr := Plan(dir, conf)
	if planErr != nil {
		return changes, planErr
	}

	if !changes {
		return false, nil
	}

	return true, terraform.Apply(dir, planName, conf.TerraformPath, conf.Timeouts.TerraformApply)
}

func Destroy(dir string, conf config.Config) error {
	initErr := terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return initErr
	}

	destroyErr := terraform.Destroy(dir, conf.TerraformPath, conf.Timeouts.TerraformDestroy)
	if destroyErr != nil {
		return destroyErr
	}

	return nil
}