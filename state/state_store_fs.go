package state

import (
	"errors"
	"fmt"
	yaml "gopkg.in/yaml.v2"
	"io/ioutil"

	"github.com/Ferlab-Ste-Justine/terracd/fs"
)

type FsConfig struct {
	Enabled bool
	Path    string `yaml:"-"`
}

func (conf *FsConfig) IsDefined() bool {
	return conf.Enabled
}

type FsStateStore struct {
	Config FsConfig
}

func (store *FsStateStore) Initialize() error {
	return nil
}

func (store *FsStateStore) Read() (State, error) {
	var st State

	stExists, stExistsErr := fs.PathExists(store.Config.Path)
	if stExistsErr != nil {
		return st, stExistsErr
	}
	if !stExists {
		return st, nil
	}

	input, err := ioutil.ReadFile(store.Config.Path)
	if err != nil {
		return st, errors.New(fmt.Sprintf("Error reading the state file: %s", err.Error()))
	}

	err = yaml.Unmarshal(input, &st)
	if err != nil {
		return st, errors.New(fmt.Sprintf("Error deserializing the state file: %s", err.Error()))
	}

	return st, nil
}

func (store *FsStateStore) Write(state State) error {
	output, err := yaml.Marshal(&state)
	if err != nil {
		return errors.New(fmt.Sprintf("Error serializing the state file: %s", err.Error()))
	}

	err = fs.EnsureContainingDirExists(store.Config.Path)
	if err != nil {
		return errors.New(fmt.Sprintf("Error creating the state file directory: %s", err.Error()))
	}

	err = ioutil.WriteFile(store.Config.Path, output, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Error writing the state file: %s", err.Error()))
	}

	return nil
}

func (store *FsStateStore) Cleanup() error {
	return nil
}