package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"time"
	yaml "gopkg.in/yaml.v2"
)

type ConfigSourceRepoAuth struct {
	SshKeyPath     string	`yaml:"ssh_key_path"`
	KnownHostsPath string	`yaml:"known_hosts_path"`
}

type ConfigSourceRepo struct {
	Url                string
	Ref                string
	Path               string
	Auth               ConfigSourceRepoAuth
	GpgPublicKeysPaths []string	`yaml:"gpg_public_keys_paths"`
}

type ConfigSource struct {
	Dir  string
	Repo ConfigSourceRepo
}

type ConfigTimeouts struct {
	TerraformInit  time.Duration `yaml:"terraform_init"`
	TerraformPlan  time.Duration `yaml:"terraform_plan"`
	TerraformApply time.Duration `yaml:"terraform_apply"`
	Wait           time.Duration
}

type Config struct {
	TerraformPath string	`yaml:"terraform_path"`
	Sources       []ConfigSource
	Timeouts      ConfigTimeouts
	Command       string
}

func getConfigFilePath() string {
	path := os.Getenv("TERRACD_CONFIG_FILE")
	if path == "" {
	  return "config.yml"
	}
	return path
}

func getConfig() (Config, error) {
	var c Config

	b, err := ioutil.ReadFile(getConfigFilePath())
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error reading the configuration file: %s", err.Error()))
	}
	err = yaml.Unmarshal(b, &c)
	if err != nil {
		return c, errors.New(fmt.Sprintf("Error parsing the configuration file: %s", err.Error()))
	}

	if c.Command == "" {
		c.Command = "apply"
	}

	if c.Command != "apply" && c.Command != "plan" && c.Command != "wait" {
		return c, errors.New("Valid command values can only be 'plan' or 'apply'")
	}

	return c, nil
}