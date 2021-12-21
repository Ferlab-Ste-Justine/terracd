package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	yaml "gopkg.in/yaml.v2"
)

type ConfigSourceRepoAuth struct {
	SshKeyPath     string	`yaml:"ssh_key_path"`
	KnownHostsPath string	`yaml:"known_hosts_path"`
}

type ConfigSourceRepo struct {
	Url  string
	Ref  string
	Path string
	Auth ConfigSourceRepoAuth
}

type ConfigSource struct {
	Dir  string
	Repo ConfigSourceRepo
}

type Config struct {
	TerraformPath string	`yaml:"terraform_path"`
	Sources       []ConfigSource
}

func getConfig() (Config, error) {
	var c Config

	b, err := ioutil.ReadFile("config.yml")
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error reading the configuration file: %s", err.Error()))
	}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error parsing the configuration file: %s", err.Error()))
	}

	return c, nil
}