package main

import (
	"ferlab/terracd/terraform"
)

func terraformApply(dir string, conf Config) error {
	initErr := terraform.Init(dir, conf.TerraformPath)
	if initErr != nil {
		return initErr
	}

	changes, planErr := terraform.Plan(dir, conf.TerraformPath)
	if planErr != nil {
		return planErr
	}

	if changes {
		applyErr := terraform.Apply(dir, conf.TerraformPath)
		if applyErr != nil {
			return applyErr
		}
	}

	return nil
}