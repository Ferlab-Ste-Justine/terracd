package terraform

import (
	"errors"
	"fmt"
	"io/ioutil"
	yaml "gopkg.in/yaml.v2"
	tfjson "github.com/hashicorp/terraform-json"
)

type ForbiddenOperation struct {
	Provider        string
	ResourceAddress string			`yaml:"resource_address"`
	Operations      tfjson.Actions
}

type ForbiddenOperationsFile struct {
	ForbiddenOperations []ForbiddenOperation	`yaml:"forbidden_operations"`
}

func GetForbiddenOperations(paths []string) ([]ForbiddenOperation, error) {
	forbiddenOps := []ForbiddenOperation{}
	
	for _, path := range paths {
		var forOpsfile ForbiddenOperationsFile

		b, err := ioutil.ReadFile(path)
		if err != nil {
			return forbiddenOps, errors.New(fmt.Sprintf("Error reading forbidden operations file %s: %s", path, err.Error()))
		}
		err = yaml.Unmarshal(b, &forOpsfile)
		if err != nil {
			return forbiddenOps, errors.New(fmt.Sprintf("Error parsing forbidden operations file %s: %s", path, err.Error()))
		}

		forbiddenOps = append(forbiddenOps, forOpsfile.ForbiddenOperations...)
	}

	return forbiddenOps, nil
}