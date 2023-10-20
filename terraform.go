package main

import (
	"os"
	"path"

	"github.com/Ferlab-Ste-Justine/terracd/config"
	"github.com/Ferlab-Ste-Justine/terracd/fs"
	"github.com/Ferlab-Ste-Justine/terracd/terraform"
)

func terraformMigrateBackend(dir string, conf config.Config) error {
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

func terraformPlan(dir string, conf config.Config) error {
	planName := "terracd-plan"
	forbiddenOpsFsPattern := "*.terracd-fo.yml"

	initErr := terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return initErr
	}

	changes, planErr := terraform.Plan(dir, planName, conf.TerraformPath, conf.Timeouts.TerraformPlan)
	if planErr != nil {
		return planErr
	}

	if changes {
		forbiddenOpsFiles, foFilesErr := fs.FindFiles(dir, forbiddenOpsFsPattern)
		if foFilesErr != nil {
			return foFilesErr
		}

		forbiddenOps, foErr := terraform.GetForbiddenOperations(forbiddenOpsFiles)
		if foErr != nil {
			return foErr
		}

		checkErr := terraform.CheckPlan(dir, planName, conf.TerraformPath, forbiddenOps)
		if checkErr != nil {
			return checkErr
		}
	}

	return nil
}

func terraformApply(dir string, conf config.Config) error {
	planName := "terracd-plan"
	forbiddenOpsFsPattern := "*.terracd-fo.yml"

	initErr := terraform.Init(dir, conf.TerraformPath, conf.Timeouts.TerraformInit)
	if initErr != nil {
		return initErr
	}

	changes, planErr := terraform.Plan(dir, planName, conf.TerraformPath, conf.Timeouts.TerraformPlan)
	if planErr != nil {
		return planErr
	}

	if changes {
		forbiddenOpsFiles, foFilesErr := fs.FindFiles(dir, forbiddenOpsFsPattern)
		if foFilesErr != nil {
			return foFilesErr
		}

		forbiddenOps, foErr := terraform.GetForbiddenOperations(forbiddenOpsFiles)
		if foErr != nil {
			return foErr
		}

		checkErr := terraform.CheckPlan(dir, planName, conf.TerraformPath, forbiddenOps)
		if checkErr != nil {
			return checkErr
		}

		applyErr := terraform.Apply(dir, planName, conf.TerraformPath, conf.Timeouts.TerraformApply)
		if applyErr != nil {
			return applyErr
		}
	}

	return nil
}
