package main

import (
	"ferlab/terracd/terraform"
	"ferlab/terracd/fs"
)

func terraformApply(dir string, conf Config) error {
	planName := "terracd-plan"
	forbiddenOpsFsPattern := "*.terracd-fo.yml"

	initErr := terraform.Init(dir, conf.TerraformPath)
	if initErr != nil {
		return initErr
	}

	changes, planErr := terraform.Plan(dir, planName, conf.TerraformPath)
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

		applyErr := terraform.Apply(dir, planName, conf.TerraformPath)
		if applyErr != nil {
			return applyErr
		}
	}

	return nil
}